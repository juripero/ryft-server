package main

import (
	"fmt"
	"io"
	"log"
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

// TODO: unsafe!!! need authentication!!!

// DeleteFilesParams query parameters for DELETE /files
// there is no actual difference between dirs and files - everything will be deleted
type DeleteFilesParams struct {
	Files []string `form:"file" json:"file"`
	Dirs  []string `form:"dir" json:"dir"`
}

// NewFilesParams query parameters for POST /files
type NewFileParams struct {
	File string `form:"file" json:"file"`
}

// GET /files method
func (s *Server) getFiles(c *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(c)

	var err error

	// parse request parameters
	params := GetFilesParams{}
	if err = c.Bind(&params); err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to parse request parameters"))
	}

	// get search engine
	engine, err := s.getSearchEngine(params.Local, nil /*no files*/)
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
	// recover from panics if any
	defer RecoverFromPanic(c)

	params := DeleteFilesParams{}
	if err := c.Bind(&params); err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "failed to parse request parameters"))
	}

	mountPoint, err := s.getMountPoint()
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to get mount point"))
	}

	result := map[string]string{}

	//log.WithField("dirs", params.Dirs).WithField("files", params.Files).Info("deleting")
	log.Printf("deleting files:%q dirs:%q", params.Files, params.Dirs)

	items := make([]string, 0, len(params.Dirs)+len(params.Files))
	items = append(items, params.Files...)
	items = append(items, params.Dirs...)

	// delete files and directories
	errs := deleteAll(mountPoint, items)
	for k, err := range errs {
		name := items[k]

		// in case of duplicate input
		// last result will be reported
		if err != nil {
			result[name] = err.Error()
		} else {
			result[name] = "OK"
		}
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

	mountPoint, err := s.getMountPoint()
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to get mount point"))
	}

	//log.WithField("file", params.File).Infof("saving new data")
	log.Printf("saving new data for %q", params.File)

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

// get mount point path from search engine
func (s *Server) getMountPoint() (string, error) {
	engine, err := s.getSearchEngine(true, []string{})
	if err != nil {
		return "", err
	}

	opts := engine.Options()
	return utils.AsString(opts["ryftone-mount"])
}

// remove directories or/and files
func deleteAll(mountPoint string, items []string) []error {
	res := make([]error, len(items))
	for k, item := range items {
		path := filepath.Join(mountPoint, item)
		res[k] = os.RemoveAll(path)
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
