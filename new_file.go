package main

import (
	"fmt"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

type NewFileParams struct {
	Path    string `form:"file" json:"file" binding:"required"`
	Content []byte `form:"content" json:"content" binding:"required"`
}

func (s *Server) newFile(c *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(c)

	var err error
	result := "Ok"

	mountPoint, _ := utils.AsString(s.BackendOptions["ryftone-mount"])

	params := NewFileParams{}
	if err := c.Bind(&params); err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "failed to parse request parameters"))
	}

	file := utils.File{
		Path:    params.Path,
		Content: params.Content,
	}
	err = utils.CreateFile(mountPoint, file)

	if err != nil {
		result = fmt.Sprintf("%s", err)
	}

	json := map[string]string{
		"result": result,
	}
	c.JSON(http.StatusOK, json)
}
