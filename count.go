package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

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
	var err error

	// parse request parameters
	params := CountParams{}
	if err = c.Bind(&params); err != nil {
		// panic(srverr.New(http.StatusBadRequest, err.Error()))
	}

	// get a new unique search index
	n := names.New()
	defer os.Remove(names.ResultsDirPath(n.IdxFile))
	defer os.Remove(names.ResultsDirPath(n.ResultFile))

	ryftParams := &RyftprimParams{
		Query:         params.Query,
		Files:         params.Files,
		Fuzziness:     params.Fuzziness,
		CaseSensitive: params.CaseSensitive,
		Nodes:         params.Nodes,
	}
	fmt.Println(">> 1 ")
	_, statistics := ryftprim(ryftParams, &n)
	fmt.Println(">> 2 ")
	stats := (<-statistics)
	//	value := m.(map[string]interface{})[matches]
	matches, err := strconv.ParseUint(fmt.Sprintf("%v", stats["matches"]), 0, 64)
	if err != nil {
		panic(srverr.New(http.StatusInternalServerError, err.Error()))
	}
	matches = matches
	c.JSON(http.StatusOK, stats)
}
