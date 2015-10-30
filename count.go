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

	c.Header("Content-Type", accept)

	// get a new unique search index
	n := names.New()

	p := progress(&params, n)

	// read an index file
	var idx *os.File
	if idx, err = crpoll.OpenFile(names.ResultsDirPath(n.IdxFile), p); err != nil {
		panic(srverr.New(http.StatusInternalServerError, err.Error()))
	}
	defer cleanup(idx)
	counter := 0
	indexes, _ := records.Poll(idx, p)
	for range indexes {
		counter++
	}
	fmt.Println()

	// c.JSON(http.StatusOK, fmt.Sprintf("Matching: %v", counter))
	c.JSON(http.StatusOK, counter)
}
