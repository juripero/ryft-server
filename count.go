package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/getryft/ryft-server/encoder"
	"github.com/getryft/ryft-server/names"
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

	_, headers := ryftprim(ryftParams, &n)
	m := <-headers
	log.Printf(" Count--- m:\n%v\n\n", m)
	setHeaders(c, m)
	// read an index file
	// var idx *os.File
	// if idx, err = crpoll.OpenFile(names.ResultsDirPath(n.IdxFile), p); err != nil {
	// 	panic(srverr.New(http.StatusInternalServerError, err.Error()))
	// }
	// defer cleanup(idx)
	// counter := uint64(0)
	// indexes, _ := records.Poll(idx, p)
	// for range indexes {
	// 	counter++
	// }
	// fmt.Println()

	// c.JSON(http.StatusOK, fmt.Sprintf("Matching: %v", counter))
	matches, err := strconv.ParseUint(fmt.Sprintf("%v", m["Matches"]), 0, 64)
	if err != nil {
		panic(srverr.New(http.StatusInternalServerError, err.Error()))
	}
	c.JSON(http.StatusOK, CountResponse{matches})
}
