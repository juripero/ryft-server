package main

import (
	"log"
	"net/http"

	"github.com/getryft/ryft-server/search/utils"
	"github.com/gin-gonic/gin"
)

// TODO: unsafe!!! need authentication!!!

// DeleteFilesParams query parameters for DELETE /files
// there is no actual difference between dirs and files - everything will be deleted
type DeleteFilesParams struct {
	Files []string `form:"file" json:"file"`
	Dirs  []string `form:"dir" json:"dir"`
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

	log.Printf("deleting dirs:%q and files:%q", params.Dirs, params.Files)

	// delete files first
	if len(params.Files) > 0 {
		errs := utils.DeleteFiles(mountPoint, params.Files)
		for k, err := range errs {
			name := params.Files[k]

			// in case of duplicate input
			// last result will be reported
			if err != nil {
				result[name] = err.Error()
			} else {
				result[name] = "OK"
			}
		}
	}

	// delete directories then
	if len(params.Dirs) > 0 {
		errs := utils.DeleteDirs(mountPoint, params.Dirs)
		for k, err := range errs {
			name := params.Dirs[k]

			// in case of duplicate input
			// last result will be reported
			if err != nil {
				result[name] = err.Error()
			} else {
				result[name] = "OK"
			}
		}
	}

	c.JSON(http.StatusOK, result)
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
