package rest

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/getryft/ryft-server/rest/codec"
	format "github.com/getryft/ryft-server/rest/format/raw"
	"github.com/getryft/ryft-server/search"
	"github.com/gin-gonic/gin"
)

// CountParams contains all the bound parameters for the /count endpoint.
type CountParams struct {
	Query    string   `form:"query" json:"query" msgpack:"query" binding:"required"`
	OldFiles []string `form:"files" json:"-" msgpack:"-"`   // obsolete: will be deleted
	Catalogs []string `form:"catalog" json:"-" msgpack:"-"` // obsolete: will be deleted
	Files    []string `form:"file" json:"files,omitempty" msgpack:"files,omitempty"`

	Mode   string `form:"mode" json:"mode,omitempty" msgpack:"mode,omitempty"`          // optional, "" for generic mode
	Width  string `form:"surrounding" json:"width,omitempty" msgpack:"width,omitempty"` // surrounding width or "line"
	Dist   uint8  `form:"fuzziness" json:"dist,omitempty" msgpack:"dist,omitempty"`     // fuzziness distance
	Case   bool   `form:"cs" json:"case,omitempty" msgpack:"case,omitempty"`            // case sensitivity flag, ES, FHS, FEDS
	Reduce bool   `form:"reduce" json:"reduce,omitempty" msgpack:"reduce,omitempty"`    // FEDS only
	Nodes  uint8  `form:"nodes" json:"nodes,omitempty" msgpack:"nodes,omitempty"`

	KeepDataAs  string `form:"data" json:"data,omitempty" msgpack:"data,omitempty"`
	KeepIndexAs string `form:"index" json:"index,omitempty" msgpack:"index,omitempty"`
	Delimiter   string `form:"delimiter" json:"delimiter,omitempty" msgpack:"delimiter,omitempty"`

	Local bool `form:"local" json:"local,omitempty" msgpack:"local,omitempty"`
}

// Handle /count endpoint.
func (server *Server) DoCount(ctx *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(ctx)

	var err error

	// parse request parameters
	params := CountParams{
		Case: true,
	}
	if err := ctx.Bind(&params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	// backward compatibility old files and catalogs (just aliases)
	params.Files = append(params.Files, params.OldFiles...)
	params.OldFiles = nil // reset
	params.Files = append(params.Files, params.Catalogs...)
	params.Catalogs = nil // reset
	if len(params.Files) == 0 {
		panic(NewError(http.StatusBadRequest,
			"no any file or catalog provided"))
	}

	accept := ctx.NegotiateFormat(codec.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = codec.MIME_JSON
	}
	if accept != codec.MIME_JSON { //if accept == encoder.MIME_MSGPACK || accept == encoder.MIME_XMSGPACK {
		panic(NewError(http.StatusUnsupportedMediaType,
			"Only JSON format is supported for now"))
	}

	// get search engine
	userName, authToken, homeDir, userTag := server.parseAuthAndHome(ctx)
	engine, err := server.getSearchEngine(params.Local, params.Files, authToken, homeDir, userTag)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get search engine"))
	}

	// prepare search configuration
	cfg := search.NewConfig(params.Query, params.Files...)
	cfg.Mode = params.Mode
	if strings.EqualFold(params.Width, "line") {
		cfg.Width = -1
	} else if v, err := strconv.ParseUint(params.Width, 10, 16); err == nil {
		cfg.Width = int(v)
	} else {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse surrounding width"))
	}
	cfg.Dist = uint(params.Dist)
	cfg.Case = params.Case
	cfg.Reduce = params.Reduce
	cfg.Nodes = uint(params.Nodes)
	cfg.KeepDataAs = params.KeepDataAs
	cfg.KeepIndexAs = params.KeepIndexAs
	cfg.Delimiter = params.Delimiter
	cfg.ReportIndex = false // /count
	cfg.ReportData = false
	// cfg.Limit = 0

	log.WithFields(map[string]interface{}{
		"config":  cfg,
		"user":    userName,
		"home":    homeDir,
		"cluster": userTag,
	}).Infof("start /count")
	res, err := engine.Search(cfg)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to start search"))
	}
	defer log.WithField("result", res).Infof("/count done")

	// in case of unexpected panic
	// we need to cancel search request
	// to prevent resource leaks
	defer func() {
		if !res.IsDone() { // cancel processing
			if errors, records := res.Cancel(); errors > 0 || records > 0 {
				log.WithFields(map[string]interface{}{
					"errors":  errors,
					"records": records,
				}).Debugf("some errors/records are ignored (panic recover)")
			}
		}
	}()

	server.onSearchStarted(cfg)
	defer server.onSearchStopped(cfg)

	// process results!
	for {
		select {
		case <-ctx.Writer.CloseNotify(): // cancel processing
			log.Warnf("cancelling by user (connection is gone)...")
			if errors, records := res.Cancel(); errors > 0 || records > 0 {
				log.WithFields(map[string]interface{}{
					"errors":  errors,
					"records": records,
				}).Debugf("some errors/records are ignored")
			}
			return // cancelled

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				// log.WithField("record", rec).Debugf("record ignored") // FIXME: DEBUG
				_ = rec // ignore records
			}

		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				// log.WithField("error", err).Debugf("error ignored") // FIXME: DEBUG
				panic(err) // TODO: check this?
			}

		case <-res.DoneChan:
			// drain the records
			for rec := range res.RecordChan {
				// log.WithField("record", rec).Debugf("*** record ignored") // FIXME: DEBUG
				_ = rec // ignore records
			}

			// ... and errors
			for err := range res.ErrorChan {
				// log.WithField("error", err).Debugf("error ignored") // FIXME: DEBUG
				panic(err) // TODO: check this?
			}

			if res.Stat != nil {
				if server.Config.ExtraRequest {
					res.Stat.Extra["request"] = &params
				}
				xstat := format.FromStat(res.Stat)
				ctx.IndentedJSON(http.StatusOK, xstat)
			} else {
				panic(NewError(http.StatusInternalServerError,
					"no search statistics available"))
			}

			return // done
		}
	}
}
