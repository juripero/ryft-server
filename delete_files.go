package main

import (
	"fmt"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DeleteFilesParams struct {
	File []string `form:"file" json:"file"`
	Dir  []string `form:"dir" json:"dir"`
}

func (s *Server) deleteFiles(c *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(c)

	var err error
	result := "Ok"

	//mountPoint := s.BackendOptions["ryftone-mount"]
	params := DeleteFilesParams{}
	if err := c.Bind(&params); err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "failed to parse request parameters"))
	}

	if len(params.Dir) > 0 {
		err = utils.DeleteDirs("/ryftone", params.Dir)
	}

	if len(params.File) > 0 {
		err = utils.DeleteFiles("/ryftone", params.File)
	}

	if err != nil {
		result = fmt.Sprintf("%s", err)
	}

	json := map[string]string{
		"result": result,
	}
	c.JSON(http.StatusOK, json)
}
