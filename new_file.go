package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/getryft/ryft-server/search/utils"
	"github.com/gin-gonic/gin"
)

// DeleteFilesParams query parameters for DELETE /files
// there is no actual difference between dirs and files - everything will be deleted
type NewFileParams struct {
	File string `form:"file" json:"file"`
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

	log.Printf("saving new data to %q\n%+v\n\n", params.File, c.Request)

	file, _, err := c.Request.FormFile("content")
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "no \"content\" form data provided"))
	}
	defer file.Close()
	path, err := utils.CreateFile(mountPoint, params.File, file)

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
