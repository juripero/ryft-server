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

//swagger:parameters count
type CountParams struct {
	// Search query, for example: ( RAW_TEXT CONTAINS "night" )
	// Required: true
	Query string `form:"query" json:"query" binding:"required"`
	// Source files
	//Required: true
	Files []string `form:"files" json:"files" binding:"required"`
	// Is the fuzziness of the search. Measured as the maximum Hamming distance.
	Fuzziness uint8 `form:"fuzziness" json:"fuzziness"`
	// Case sensitive flag
	CaseSensitive bool `form:"cs" json:"cs"`
	//Active Nodes Count
	//minimum: 0
	//maximum: 4
	Nodes uint8 `form:"nodes" json:"nodes"`
}

func count(c *gin.Context) {
	defer srverr.DeferRecover(c)

	var err error

	// parse request parameters
	params := NewSearchParams()
	if err = c.Bind(&params); err != nil {
		// panic(srverr.New(http.StatusBadRequest, err.Error()))
	}

	accept := c.NegotiateFormat(encoder.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = encoder.MIMEJSON
	}

	setHeaders(c)

	c.Header("Content-Type", accept)
	// get a new unique search index
	n := names.New()

	p := ryftprim(&params, &n)

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
	c.JSON(http.StatusOK, struct {
		Matches uint64
	}{counter})
}
