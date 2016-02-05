package main

import (
	"log"
	"net/http"
	"net/url"

	"github.com/getryft/ryft-server/codec"
	"github.com/getryft/ryft-server/search"
	"github.com/gin-gonic/gin"
)

// CountParams is a parameters for matches count endpoint
type CountParams struct {
	Query         string   `form:"query" json:"query" binding:"required"`
	Files         []string `form:"files" json:"files" binding:"required"`
	Fuzziness     uint8    `form:"fuzziness" json:"fuzziness"`
	CaseSensitive bool     `form:"cs" json:"cs"`
	Nodes         uint8    `form:"nodes" json:"nodes"`
	Local         bool     `form:"local" json:"local"`
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
	engine, err := s.getSearchEngine(params.Local)
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
	cfg.Surrounding = 0
	cfg.Fuzziness = uint(params.Fuzziness)
	cfg.CaseSensitive = params.CaseSensitive
	cfg.Nodes = uint(params.Nodes)

	res, err := engine.Count(cfg)
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to start search"))
	}

	// TODO: for cloud code get other ryftprim.Result objects and merge together
	// [[[ ]]]]

	for {
		select {
		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				log.Printf("REC: %s", rec)
				// ignore records
			}

		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				log.Printf("ERR: %s", err)
				// TODO: report error
			}

		case <-res.DoneChan:
			log.Printf("DONE: %s", res.Stat)
			s := map[string]interface{}{
				"matches":    res.Stat.Matches,
				"totalBytes": res.Stat.TotalBytes,
				"duration":   res.Stat.Duration,
			}
			ctx.JSON(http.StatusOK, s)
			return
		}
	}
}
