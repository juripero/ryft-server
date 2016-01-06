package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/getryft/ryft-server/crpoll"
	"github.com/getryft/ryft-server/encoder"
	"github.com/getryft/ryft-server/names"
	"github.com/getryft/ryft-server/records"
	"github.com/getryft/ryft-server/srverr"
	"github.com/gin-gonic/gin"
)

// CountParams is a parameters for matches count endpoint
type CountParams struct {
	Query         string   `form:"query" json:"query" binding:"required"`
	Files         []string `form:"files" json:"files" binding:"required"`
	Fuzziness     uint8    `form:"fuzziness" json:"fuzziness"`
	CaseSensitive bool     `form:"cs" json:"cs"`
	Nodes         uint8    `form:"nodes" json:"nodes"`
}

// CountResponse returnes matches for query
//
type CountResponse struct {
	Mathces uint64 `json:"matches, string"`
}

func count(c *gin.Context) {
	defer srverr.DeferRecover(c)

	var err error

	// parse request parameters
	params := CountParams{}
	if err = c.Bind(&params); err != nil {
		// panic(srverr.New(http.StatusBadRequest, err.Error()))
	}

	accept := c.NegotiateFormat(encoder.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = encoder.MIMEJSON
	}

	c.Header("Content-Type", accept)
	// get a new unique search index
	n := names.New()

	ryftParams := &RyftprimParams{
		Query:         params.Query,
		Files:         params.Files,
		Fuzziness:     params.Fuzziness,
		CaseSensitive: params.CaseSensitive,
		Nodes:         params.Nodes,
	}

	p := ryftprim(ryftParams, &n)

	// read an index file
	var idx *os.File
	if idx, err = crpoll.OpenFile(names.ResultsDirPath(n.IdxFile), p); err != nil {
		panic(srverr.New(http.StatusInternalServerError, err.Error()))
	}
	defer cleanup(idx)
	counter := uint64(0)
	indexes, _ := records.Poll(idx, p)
	for range indexes {
		counter++
	}
	fmt.Println()

	// c.JSON(http.StatusOK, fmt.Sprintf("Matching: %v", counter))
	c.JSON(http.StatusOK, CountResponse{counter})
}
