package main

import (
	"net/http"

	"github.com/getryft/ryft-server/encoder"
	"github.com/getryft/ryft-server/srverr"
	"github.com/gin-gonic/gin"
)

type FilesParams struct {
	Dir   string `form:"dir" json:"dir"`
	Local bool   `form:"local" json:"local"`
}

func (s *Server) files(c *gin.Context) {
	// recover from panics if any
	defer srverr.Recover(c)

	var err error

	// parse request parameters
	params := FilesParams{}
	if err = c.Bind(&params); err != nil {
		panic(srverr.NewWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to parse request parameters"))
	}

	// get search engine
	engine, err := s.getSearchEngine(params.Local)
	if err != nil {
		panic(srverr.NewWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to get search engine"))
	}

	accept := c.NegotiateFormat(encoder.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = encoder.MIME_JSON
	}
	if accept != encoder.MIME_JSON { //if accept == encoder.MIME_MSGPACK || accept == encoder.MIME_XMSGPACK {
		panic(srverr.New(http.StatusUnsupportedMediaType,
			"Only JSON format is supported for now"))
	}

	info, err := engine.Files(params.Dir)
	if err != nil {
		// TODO: detail description?
		panic(srverr.New(http.StatusNotFound, err.Error()))
	}

	// TODO: use transcoder/dedicated structure instead of simple map!
	json := map[string]interface{}{
		"dir":     info.Path,
		"files":   info.Files,
		"folders": info.Dirs,
	}
	c.JSON(http.StatusOK, json)
}