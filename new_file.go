package main

import (
	"fmt"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Server) newFile(c *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(c)

	var err error
	result := "Ok"

	mountPoint, _ := utils.AsString(s.BackendOptions["ryftone-mount"])

	file := bindParams(c)
	fmt.Println("f", file)
	path, err := utils.CreateFile(mountPoint, file)

	if err != nil {
		result = fmt.Sprintf("%s", err)
	}

	json := map[string]string{
		"path":   path,
		"result": result,
	}
	c.JSON(http.StatusOK, json)
}

func bindParams(c *gin.Context) utils.File {
	file, _, _ := c.Request.FormFile("content")
	path := c.Query("file")

	return utils.File{
		Path:   path,
		Reader: file,
	}
}
