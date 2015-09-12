// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/getryft/ryft-rest-api/ryft-server/binding"
	"github.com/getryft/ryft-rest-api/ryft-server/crpoll"
	"github.com/getryft/ryft-rest-api/ryft-server/names"
	"github.com/getryft/ryft-rest-api/ryft-server/outstream"
	"github.com/getryft/ryft-rest-api/ryft-server/progress"
	"github.com/getryft/ryft-rest-api/ryft-server/records"
	"github.com/getryft/ryft-rest-api/ryft-server/srverr"
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

	log.Printf("** start binding")
	var s *binding.Search
	var err error
	if s, err = binding.NewSearch(c); err != nil {
		panic(srverr.New(http.StatusBadRequest, err.Error()))
	}

	c.Header("Content-Type", gin.MIMEPlain)
	if s.IsOutJson() {
		c.Header("Content-Type", gin.MIMEPlain)
	} else if s.IsOutMsgpk() {
		c.Header("Content-Type", "application/x-msgpack")
	} else {
		panic(srverr.New(http.StatusBadRequest, "Supported formats (Content-Type): application/json, application/x-msgpack"))
	}

	n := names.New()
	log.Printf("SEARCH(%d): %s", n.Index, c.Request.URL.String())

	p := progress.Progress(s, n)

	var idx, res *os.File
	if idx, err = crpoll.OpenFile(names.ResultsDirPath(n.IdxFile), p); err != nil {
		panic(srverr.New(http.StatusInternalServerError, err.Error()))
	}

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
	
		switch s.State {
			case binding.StateBegin:
				log.Println("StateBegin") 
				s.State = binding.StateBody
			case binding.StateBody:
				log.Println("StateBody")
				s.State = binding.StateEnd 
			case binding.StateEnd:
				log.Println("StateEnd") 
		}
	
		var err error
		err = outstream.Write(s, recs, res, w, drop)
		if err != nil {

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



func main() {
	log.SetFlags(log.Lmicroseconds)
	readParameters()

	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))


	r.GET("/search", search)
	r.StaticFile("/", "./index.html");

	// Clean previously created folder
	if err := os.RemoveAll(names.ResultsDirPath()); err != nil {
		log.Printf("Could not delete %s with error %s", names.ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	// Create folder for results cache
	if err := os.MkdirAll(names.ResultsDirPath(), 0777); err != nil {
		log.Printf("Could not create directory %s with error %s", names.ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	// Name Generator will produce unique file names for each new results files
	names.StartNamesGenerator()
	r.Run(fmt.Sprintf(":%d", names.Port))

}

// https://golang.org/src/net/http/status.go -- statuses
