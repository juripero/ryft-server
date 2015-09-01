// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/DataArt/ryft-rest-api/ryft-server/binding"
	"github.com/DataArt/ryft-rest-api/ryft-server/crpoll"
	"github.com/DataArt/ryft-rest-api/ryft-server/jsonstream"
	"github.com/DataArt/ryft-rest-api/ryft-server/names"
	"github.com/DataArt/ryft-rest-api/ryft-server/progress"
	"github.com/DataArt/ryft-rest-api/ryft-server/records"
	"github.com/DataArt/ryft-rest-api/ryft-server/srverr"

	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
)

var (
	KeepResults = false
)

func readParameters() {
	portPtr := flag.Int("port", 8765, "The port of the REST-server")
	keepResultsPtr := flag.Bool("keep-results", false, "Keep results or delete after response")

	flag.Parse()

	names.Port = *portPtr
	KeepResults = *keepResultsPtr
}

func search(c *gin.Context) {
	defer srverr.DeferRecover(c)

	var s *binding.Search
	var err error
	if s, err = binding.NewSearch(c); err != nil {
		panic(srverr.New(http.StatusBadRequest, err.Error()))
	}

	n := names.New()
	log.Printf("SEARCH(%d): %s", n.Index, c.Request.URL.String())

	p := progress.Progress(s, n)

	var idx, res *os.File
	if idx, err = crpoll.OpenFile(names.ResultsDirPath(n.IdxFile), p); err != nil {
		panic(srverr.New(http.StatusInternalServerError, err.Error()))
	}
	log.Printf("%d: idx-file opened", n.Index)

	defer func() {
		if idx != nil {
			idx.Close()
			if !KeepResults {
				os.Remove(idx.Name())
			}
		}
	}()

	if res, err = crpoll.OpenFile(names.ResultsDirPath(n.ResultFile), p); err != nil {
		panic(srverr.New(http.StatusInternalServerError, err.Error()))
	}
	log.Printf("%d: res-file opened", n.Index)

	defer func() {
		if res != nil {
			res.Close()
			if !KeepResults {
				os.Remove(res.Name())
			}
		}
	}()

	recs, drop := records.Poll(idx, p)

	c.Stream(func(w io.Writer) bool {
		if err := jsonstream.Write(recs, res, w, drop); err != nil {
			idx.Close()
			idx = nil
			if !KeepResults {
				os.Remove(names.ResultsDirPath(n.IdxFile))
			}
			res.Close()
			res = nil
			if !KeepResults {
				os.Remove(names.ResultsDirPath(n.ResultFile))
			}
		}
		return false
	})
}

func testOk(c *gin.Context) {
	defer srverr.DeferRecover(c)

	c.Stream(func(w io.Writer) bool {

		w.Write([]byte("["))
		firstIteration := true
		for i := 0; i <= 100; i++ {
			if !firstIteration {
				w.Write([]byte(","))
			}

			record := gin.H{"number": i}
			bytes, err := json.Marshal(record)
			if err != nil {
				panic(srverr.New(http.StatusInternalServerError, err.Error()))
			}

			w.Write(bytes)

			firstIteration = false
		}

		w.Write([]byte("]"))
		return false
	})
}

func main() {
	readParameters()

	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	indexTemplate := template.Must(template.New("index").Parse(IndexHTML))
	r.SetHTMLTemplate(indexTemplate)

	r.GET("/", func(c *gin.Context) {
		defer srverr.DeferRecover(c)
		c.HTML(http.StatusOK, "index", nil)
	})

	r.GET("/search/test-ok", func(c *gin.Context) {
		c.Header("Content-Type", gin.MIMEPlain)
		testOk(c)
	})

	r.GET("/searchtest", func(c *gin.Context) {
		c.Header("Content-Type", gin.MIMEPlain)
		searchtest(c)
	})

	r.GET("/search/test-fail", func(c *gin.Context) {
		defer srverr.DeferRecover(c)
		panic(srverr.New(http.StatusInternalServerError, "Test error"))
	})

	r.GET("/search", func(c *gin.Context) {
		c.Header("Content-Type", gin.MIMEPlain)
		search(c)
	})

	if err := os.RemoveAll(names.ResultsDirPath()); err != nil {
		log.Printf("Could not delete %s with error %s", names.ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	if err := os.MkdirAll(names.ResultsDirPath(), 0777); err != nil {
		log.Printf("Could not create directory %s with error %s", names.ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	names.StartNamesGenerator()
	log.SetFlags(log.Ltime)

	r.Run(fmt.Sprintf(":%d", names.Port))

}

// https://golang.org/src/net/http/status.go -- statuses
