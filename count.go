package main

import (
	"net/http"
	"net/url"
	"os"

	"github.com/getryft/ryft-server/names"
	"github.com/getryft/ryft-server/ryftprim"
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

	defer srverr.Recover(c)

	// parse request parameters
	params := CountParams{}
	if err := c.Bind(&params); err != nil {
		panic(srverr.New(http.StatusInternalServerError, err.Error()))
	}

	// get a new unique search index
	n := names.New()
	defer os.Remove(names.ResultsDirPath(n.IdxFile))
	defer os.Remove(names.ResultsDirPath(n.ResultFile))

	query, aErr := url.QueryUnescape(params.Query)

	if aErr != nil {
		panic(srverr.New(http.StatusBadRequest, aErr.Error()))
	}

	results := ryftprim.Search(&ryftprim.Params{
		Query:         query,
		Files:         params.Files,
		Fuzziness:     params.Fuzziness,
		CaseSensitive: params.CaseSensitive,
		Nodes:         params.Nodes,
		ResultsFile:   n.FullResultsPath(),
	})

	// TODO: for cloud code get other ryftprim.Result objects and merge together
	// [[[ ]]]]

	select {
	case stats := <-results.Stats:
		c.JSON(http.StatusOK, stats)

	case err := <-results.Errors:
		//		c.AbortWithError(http.StatusInternalServerError, err)
		panic(err)
	}
}
