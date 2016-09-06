package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/getryft/ryft-server/codec"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/gin-gonic/gin"
)

// GetFileParams query parameters for GET /files
type GetFilesParams struct {
	Dir   string `form:"dir" json:"dir"`
	Local bool   `form:"local" json:"local"`
}

// DeleteFilesParams query parameters for DELETE /files
// there is no actual difference between dirs and files - everything will be deleted
type DeleteFilesParams struct {
	Files []string `form:"file" json:"file"`
	Dirs  []string `form:"dir" json:"dir"`
}

// NewFilesParams query parameters for POST /files
type NewFilesParams struct {
	File    string `form:"file" json:"file"`
	Catalog string `form:"catalog" json:"catalog"`
	// TODO: catalog options
}

// GET /files method
func (s *Server) getFiles(c *gin.Context) {
	defer RecoverFromPanic(c)

	// parse request parameters
	params := GetFilesParams{}
	if err := c.Bind(&params); err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "failed to parse request parameters"))
	}

	// get search engine
	userName, authToken, homeDir, userTag := s.parseAuthAndHome(c)
	engine, err := s.getSearchEngine(params.Local, nil /*no files*/, authToken, homeDir, userTag)
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to get search engine"))
	}

	accept := c.NegotiateFormat(codec.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = codec.MIME_JSON
	}
	if accept != codec.MIME_JSON { //if accept == encoder.MIME_MSGPACK || accept == encoder.MIME_XMSGPACK {
		panic(NewServerError(http.StatusUnsupportedMediaType,
			"Only JSON format is supported for now"))
	}

	log.WithField("dir", params.Dir).WithField("user", userName).
		WithField("home", homeDir).WithField("cluster", userTag).
		Infof("start /files")
	info, err := engine.Files(params.Dir)
	if err != nil {
		// TODO: detail description?
		panic(NewServerError(http.StatusNotFound, err.Error()))
	}

	// TODO: if params.Sort {
	// sort names in the ascending order
	sort.Strings(info.Files)
	sort.Strings(info.Dirs)

	// TODO: use transcoder/dedicated structure instead of simple map!
	json := map[string]interface{}{
		"dir":     info.Path,
		"files":   info.Files,
		"folders": info.Dirs,
	}
	c.JSON(http.StatusOK, json)
}

// DELETE /files method
/* to test method:
curl -X DELETE -s "http://localhost:8765/files?file=p*.txt" | jq .
*/
func (s *Server) deleteFiles(c *gin.Context) {
	defer RecoverFromPanic(c)

	// parse request parameters
	params := DeleteFilesParams{}
	if err := c.Bind(&params); err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "failed to parse request parameters"))
	}

	userName, _, homeDir, _ := s.parseAuthAndHome(c)
	mountPoint, err := s.getMountPoint(homeDir)
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to get mount point"))
	}

	log.WithField("dirs", params.Dirs).
		WithField("files", params.Files).
		WithField("user", userName).
		WithField("home", homeDir).
		Info("deleting...")

	result := map[string]string{}
	updateResult := func(name string, err error) {
		// in case of duplicate input
		// last result will be reported
		if err != nil {
			result[name] = err.Error()
		} else {
			result[name] = "OK" // "DELETED"
		}
	}

	// delete directories first ...
	for dir, err := range deleteAll(mountPoint, params.Dirs) {
		updateResult(dir, err)
	}

	// ... then delete files
	for file, err := range deleteAll(mountPoint, params.Files) {
		updateResult(file, err)
	}

	c.JSON(http.StatusOK, result)
}

// POST /files method
/* to test method:
curl -X POST -F content=@/path/to/file.txt -s "http://localhost:8765/files?file=/test/file\{\{random\}\}.txt" | jq .
curl -X POST --data "hello" -H 'Content-Type: application/octet-stream' -s "http://localhost:8765/files?file=/test/file\{\{random\}\}.txt" | jq .
*/
func (s *Server) newFiles(c *gin.Context) {
	defer RecoverFromPanic(c)

	// parse request parameters
	params := NewFilesParams{}
	if err := c.Bind(&params); err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "failed to parse request parameters"))
	}

	// get name from "filename" form value if it's not provided in query
	if fn := c.Request.FormValue("filename"); len(fn) != 0 && len(params.File) == 0 {
		params.File = fn
	}

	if len(params.File) == 0 {
		panic(NewServerError(http.StatusBadRequest,
			"no valid filename provided"))
	}

	catalog := randomizePath(params.Catalog)

	userName, _, homeDir, _ := s.parseAuthAndHome(c)
	mountPoint, err := s.getMountPoint(homeDir)
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to get mount point"))
	}

	var file io.Reader
	var fileLen int64 // -1 unknown

	contentType := c.ContentType()
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "multipart/form-data":
		f, _, err := c.Request.FormFile("content")
		if err != nil {
			panic(NewServerErrorWithDetails(http.StatusBadRequest,
				err.Error(), `no "content" form data provided`))
		}
		defer f.Close()
		file = f
		fileLen = -1 // c.Request.ContentLength
		log.WithField("length", fileLen).Debugf("saving multipart form data...")

	case "application/octet-stream":
		file = c.Request.Body
		fileLen = c.Request.ContentLength
		log.WithField("length", fileLen).Debugf("saving octet-stream...")

	default:
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			contentType, "unexpected content type"))
	}

	response := map[string]interface{}{}
	log.WithField("file", params.File).
		WithField("user", userName).
		WithField("home", homeDir).
		WithField("catalog", catalog).
		Infof("saving new data...")

	if len(catalog) != 0 {
		filename := randomizePath(params.File)
		path, offset, length, err := updateCatalog(mountPoint, catalog, filename, file, fileLen)

		if err != nil {
			response["error"] = err.Error()
		} else {
			response["catalog"] = catalog
			response["data_path"] = path // data path to search
			response["data_pos"] = offset
			response["length"] = length // not total, just this part
		}
	} else { // regular file
		path, length, err := createFile(mountPoint, params.File, file)

		if err != nil {
			response["error"] = err.Error()
			// TODO: use dedicated HTTP status code
		} else {
			response["path"] = path
			response["length"] = length
		}
	}

	c.JSON(http.StatusOK, response)
}

// get mount point path from local search engine
func (s *Server) getMountPoint(homeDir string) (string, error) {
	engine, err := s.getLocalSearchEngine(homeDir)
	if err != nil {
		return "", err
	}

	opts := engine.Options()
	return utils.AsString(opts["ryftone-mount"])
}

// remove directories or/and files
func deleteAll(mountPoint string, items []string) map[string]error {
	res := map[string]error{}
	for _, item := range items {
		path := filepath.Join(mountPoint, item)
		matches, err := filepath.Glob(path)
		if err != nil {
			res[item] = err
			continue
		}

		// remove all matches
		for _, file := range matches {
			rel, err := filepath.Rel(mountPoint, file)
			if err != nil {
				rel = file // ignore error and get absolute path
			}
			res[rel] = os.RemoveAll(file)
		}
	}

	return res
}

// createFile creates new file.
// Unique file name could be generated if path contains special keywords.
// Returns generated path, length and error if any.
func createFile(mountPoint string, path string, content io.Reader) (string, uint64, error) {
	rbase := randomizePath(path) // first replace all {{random}} tokens
	rpath := rbase

	// create all parent directories
	pdir := filepath.Join(mountPoint, filepath.Dir(rpath))
	err := os.MkdirAll(pdir, 0755)
	if err != nil {
		return rpath, 0, err
	}

	// try to create file, if file already exists try with updated name
	for k := 0; ; k++ {
		fullpath := filepath.Join(mountPoint, rpath)
		f, err := os.OpenFile(fullpath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			if path != rbase && os.IsExist(err) {
				// generate new unique name
				ext := filepath.Ext(rbase)
				base := strings.TrimSuffix(rbase, ext)
				rpath = fmt.Sprintf("%s-%d%s", base, k+1, ext)

				continue
			}
			return rpath, 0, err
		}
		defer f.Close()

		// copy the file content
		w, err := io.Copy(f, content)
		if err != nil {
			log.WithError(err).WithField("file", rpath).
				Warnf("failed to save data")

			// do not leave partially saved data!
			_ = os.RemoveAll(fullpath)

			return rpath, 0, err
		}

		// return path to file without mountpoint
		return rpath, uint64(w), nil // OK
	}
}

// replace {{random}} sections of filename with random token.
// random token is based on current unix time in nanoseconds.
// multiple {{random}} are possible
func randomizePath(path string) string {
	token := func(string) string {
		return fmt.Sprintf("%016x", time.Now().UnixNano())
	}

	// TODO: use some hash here

	re := regexp.MustCompile(`{{random}}`)
	return re.ReplaceAllStringFunc(path, token)
}
