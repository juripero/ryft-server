package main

import (
	"net/http"
	"sort"

	"github.com/getryft/ryft-server/codec"
	"github.com/gin-gonic/gin"
)

type FilesParams struct {
	Dir   string `form:"dir" json:"dir"`
	Local bool   `form:"local" json:"local"`
}

func (s *Server) files(c *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(c)

	var err error

	// parse request parameters
	params := FilesParams{}
	if err = c.Bind(&params); err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to parse request parameters"))
	}

	// get search engine
	authToken, homeDir, userTag := s.parseAuthAndHome(c)
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
