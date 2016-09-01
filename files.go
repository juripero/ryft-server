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
type NewFileParams struct {
	File string `form:"file" json:"file"`
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
func (s *Server) newFile(c *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(c)

	params := NewFileParams{}
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

	log.WithField("file", params.File).
		WithField("user", userName).
		WithField("home", homeDir).
		Infof("saving new data")

	file, _, err := c.Request.FormFile("content")
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "no \"content\" form data provided"))
	}
	defer file.Close()
	path, err := createFile(mountPoint, params.File, file)

	var result string
	if err != nil {
		result = fmt.Sprintf("%s", err)
		// TODO: use dedicated HTTP status code
	} else {
		log.Printf("saved to %q", path)
		result = "OK"
	}

	response := map[string]string{
		"path":   path,
		"result": result,
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
// Returns generated path and error if any.
func createFile(mountPoint string, path string, content io.Reader) (string, error) {
	rbase := randomizePath(path) // first replace all <random> tokens
	rpath := rbase

	// create all parent directories
	pdir := filepath.Join(mountPoint, filepath.Dir(rpath))
	err := os.MkdirAll(pdir, 0755)
	if err != nil {
		return rpath, err
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
			return rpath, err
		}
		defer f.Close()

		// copy the file content
		_, err = io.Copy(f, content)
		if err != nil {
			// TODO: remove corrupted file?
			return rpath, err
		}

		// return path to file without mountpoint
		return rpath, nil // OK
	}
}

// replace <random> sections of filename with random token.
// random token is based on current unix time in nanoseconds.
// multiple <random> are possible
func randomizePath(path string) string {
	token := func(string) string {
		return fmt.Sprintf("%016x", time.Now().UnixNano())
	}

	re := regexp.MustCompile("<random>")
	return re.ReplaceAllStringFunc(path, token)
}
