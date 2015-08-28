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
	//Port        = 8765  //command line "port"
	KeepResults = false //command line "keep-results"
)

// var Observer *fsobserver.Observer

func readParameters() {
	portPtr := flag.Int("port", 8765, "The port of the REST-server")
	keepResultsPtr := flag.Bool("keep-results", false, "Keep results or delete after response")

	flag.Parse()

	//Port = *portPtr
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
		if err := jsonstream.Write(recs, res, w, drop); err != nil {
			idx.Close()
			idx = nil
			if !KeepResults {
				os.Remove(idx.Name())
			}
			res.Close()
			res = nil
			if !KeepResults {
				os.Remove(res.Name())
			}
		}
		return false
	})
}

// func searchOld(c *gin.Context) {
// 	defer deferRecover(c)

// 	s, err := binding.NewSearch(c)
// 	if err != nil {
// 		panic(&ServerError{http.StatusBadRequest, err.Error()})
// 	}

// 	n := GetNewNames()
// 	ch := make(chan error, 1)

// 	log.Printf("request: start waiting for files %+v", n)
// 	idx, res, idxops, resops := startAndWaitFiles(s, n, ch)
// 	defer func() {
// 		Observer.Unfollow(idx.Name())
// 		Observer.Unfollow(res.Name())

// 		if !KeepResults {
// 			os.Remove(idx.Name())
// 			os.Remove(res.Name())
// 			log.Println("request: file deleted")
// 		}

// 		idx.Close()
// 		res.Close()
// 		log.Println("request: ops & files closed")
// 	}()
// 	log.Println("request: all files created & opened")

// 	dropper := make(chan struct{}, 1)
// 	records := GetRecordsChan(idx, idxops, ch, dropper)

// 	c.Stream(func(w io.Writer) bool {
// 		err := generateJson(records, res, resops, w, dropper)
// 		log.Println("request: after generateJson")

// 		if err != nil {
// 			Observer.Unfollow(idx.Name())
// 			Observer.Unfollow(res.Name())
// 			idx.Close()
// 			res.Close()
// 			idx = nil
// 			res = nil
// 			log.Println("request: ops & files closed")
// 		}
// 		return false
// 	})
// 	log.Println("request: end")
// }

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

	indexTemplate := template.Must(template.New("index").Parse(IndexHTML))
	r.SetHTMLTemplate(indexTemplate)

	r.GET("/", func(c *gin.Context) {
		defer srverr.DeferRecover(c)
		c.HTML(http.StatusOK, "index", nil)
	})

	r.GET("/search/test-ok", func(c *gin.Context) {
		testOk(c)
	})

	r.GET("/search/test-fail", func(c *gin.Context) {
		defer srverr.DeferRecover(c)
		panic(srverr.New(http.StatusInternalServerError, "Test error"))
	})

	r.GET("/search", func(c *gin.Context) {
		search(c)
	})

	compressed := r.Group("/gzip")

	compressed.Use(gzip.Gzip(gzip.DefaultCompression))
	{
		compressed.GET("/", func(c *gin.Context) {
			defer srverr.DeferRecover(c)
			c.HTML(http.StatusOK, "index", nil)
		})

		compressed.GET("/search/test-ok", func(c *gin.Context) {
			c.Header("Content-Type", gin.MIMEPlain)
			testOk(c)
		})

		compressed.GET("/search/test-fail", func(c *gin.Context) {
			defer srverr.DeferRecover(c)
			panic(srverr.New(http.StatusInternalServerError, "Test error"))
		})

		compressed.GET("/search", func(c *gin.Context) {
			c.Header("Content-Type", gin.MIMEPlain)
			search(c)
		})
	}

	if err := os.RemoveAll(names.ResultsDirPath()); err != nil {
		log.Printf("Could not delete %s with error %s", names.ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	if err := os.MkdirAll(names.ResultsDirPath(), 0777); err != nil {
		log.Printf("Could not create directory %s with error %s", names.ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	// var err error
	// if Observer, err = fsobserver.NewObserver(ResultsDirPath()); err != nil {
	// 	log.Printf("Could not create directory %s observer with error %s", ResultsDirPath(), err.Error())
	// 	os.Exit(1)
	// }

	names.StartNamesGenerator()
	log.SetFlags(log.Ltime)

	r.Run(fmt.Sprintf(":%d", names.Port))

}

// https://golang.org/src/net/http/status.go -- statuses
