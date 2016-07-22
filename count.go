package main

import (
	"net/http"
	"net/url"

	"github.com/getryft/ryft-server/codec"
	format "github.com/getryft/ryft-server/format/raw"
	"github.com/getryft/ryft-server/search"
	"github.com/gin-gonic/gin"
)

// CountParams is a parameters for matches count endpoint
type CountParams struct {
	Query         string   `form:"query" json:"query" binding:"required"`
	Files         []string `form:"files" json:"files" binding:"required"`
	Mode          string   `form:"mode" json:"mode"`
	Surrounding   uint16   `form:"surrounding" json:"surrounding"`
	Fuzziness     uint8    `form:"fuzziness" json:"fuzziness"`
	CaseSensitive bool     `form:"cs" json:"cs"`
	Nodes         uint8    `form:"nodes" json:"nodes"`
	Local         bool     `form:"local" json:"local"`
	KeepDataAs    string   `form:"data" json:"data"`
	KeepIndexAs   string   `form:"index" json:"index"`
}

// CountResponse returnes matches for query
//
type CountResponse struct {
	Mathces uint64 `json:"matches, string"`
}

// Handle /count endpoint.
func (s *Server) count(ctx *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(ctx)

	var err error

	// parse request parameters
	params := CountParams{}
	if err := ctx.Bind(&params); err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to parse request parameters"))
	}

	accept := ctx.NegotiateFormat(codec.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = codec.MIME_JSON
	}
	if accept != codec.MIME_JSON { //if accept == encoder.MIME_MSGPACK || accept == encoder.MIME_XMSGPACK {
		panic(NewServerError(http.StatusUnsupportedMediaType,
			"Only JSON format is supported for now"))
	}

	// get search engine
	userName, authToken, homeDir, userTag := s.parseAuthAndHome(ctx)
	engine, err := s.getSearchEngine(params.Local, params.Files, authToken, homeDir, userTag)
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to get search engine"))
	}

	// search configuration
	cfg := search.NewEmptyConfig()
	if q, err := url.QueryUnescape(params.Query); err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "failed to unescape query"))
	} else {
		cfg.Query = q
	}
	cfg.AddFiles(params.Files) // TODO: unescape?
	cfg.Mode = params.Mode
	cfg.Surrounding = uint(params.Surrounding)
	cfg.Fuzziness = uint(params.Fuzziness)
	cfg.CaseSensitive = params.CaseSensitive
	cfg.Nodes = uint(params.Nodes)
	cfg.KeepDataAs = params.KeepDataAs
	cfg.KeepIndexAs = params.KeepIndexAs

	log.WithField("config", cfg).WithField("user", userName).
		WithField("home", homeDir).WithField("cluster", userTag).
		Infof("start /count")
	res, err := engine.Count(cfg)
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to start search"))
	}

	s.onSearchStarted(cfg)
	defer s.onSearchStopped(cfg)

	for {
		select {
		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				log.WithField("record", rec).Debugf("record ignored")
				// ignore records
			}

		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				log.WithField("error", err).Debugf("error ignored")
				// TODO: report error
			}

		case <-res.DoneChan:
			log.WithField("result", res).Infof("/count done")

			if res.Stat != nil {
				stat := format.FromStat(res.Stat)
				ctx.JSON(http.StatusOK, stat)
			} else {
				panic(NewServerError(http.StatusInternalServerError,
					"no search statistics available"))
			}

			return
		}
	}
}
