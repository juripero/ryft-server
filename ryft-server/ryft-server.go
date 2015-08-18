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
	"path/filepath"

	"github.com/DataArt/ryft-rest-api/fsobserver"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Search struct {
	Query       string `form:"query" json:"query" binding:"required"`             // For example: ( RAW_TEXT CONTAINS "night" )
	Files       string `form:"files" json:"files" binding:"required"`             // Splitted OS-specific ListSeparator: "/a/b/c:/usr/bin/file" -> "/a/b/c", "/usr/bin/file"
	Surrounding uint16 `form:"surrounding" json:"surrounding" binding:"required"` // Specifies the number of characters before the match and after the match that will be returned when the input specifier type is raw text
	Fuzziness   uint8  `form:"fuzziness" json:"fuzziness"`                        // Is the fuzziness of the search. Measured as the maximum Hamming distance.

	ExtractedFiles []string `json:"extractedFiles"` // Contains files from Files (after ExtractFiles())
}

func (s *Search) ExtractFiles() {
	s.ExtractedFiles = filepath.SplitList(s.Files)
}

var (
	Port        = 8765  //command line "port"
	KeepResults = false //command line "keep-results"
)

var Observer *fsobserver.Observer

func readParameters() {
	portPtr := flag.Int("port", 8765, "The port of the REST-server")
	keepResultsPtr := flag.Bool("keep-results", false, "Keep results or delete after response")

	flag.Parse()

	Port = *portPtr
	KeepResults = *keepResultsPtr
}

func main() {
	readParameters()

	r := gin.Default()

	indexTemplate := template.Must(template.New("index").Parse(IndexHTML))
	r.SetHTMLTemplate(indexTemplate)

	r.GET("/", func(c *gin.Context) {
		defer deferRecover(c)
		c.HTML(http.StatusOK, "index", nil)
	})

	r.GET("/search/test-ok", func(c *gin.Context) {
		defer deferRecover(c)
		c.Writer.Header()["Content-Type"] = []string{binding.MIMEJSON}
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
					panic(&ServerError{http.StatusInternalServerError, err.Error()})
				}

				w.Write(bytes)

				firstIteration = false
			}

			w.Write([]byte("]"))
			return false
		})

	})

	r.GET("/search/test-fail", func(c *gin.Context) {
		defer deferRecover(c)
		panic(&ServerError{http.StatusInternalServerError, "Test error"})
	})

	r.GET("/search", func(c *gin.Context) {
		defer deferRecover(c)

		s := new(Search)
		if err := c.Bind(s); err != nil {
			panic(&ServerError{http.StatusBadRequest, err.Error()})
		}

		s.ExtractFiles()

		n := GetNewNames()
		ch := make(chan error)

		idx, res, idxops, resops := startAndWaitFiles(s, n, ch)
		defer Observer.Unfollow(idx.Name())
		defer Observer.Unfollow(res.Name())
		_ = idxops
		_ = resops

		log.Println("response: start streaming")
		c.Stream(func(w io.Writer) bool {
			//streamJson(idx, res, w, watcher, ch)
			log.Println("response: start streaming (in)")
			w.Write([]byte("[]"))
			log.Println("response: written (in)")

			w.Close()
			return false
		})
		log.Println("response: complete streaming")

		if !KeepResults {
			// os.Remove(idx.Name())
			// os.Remove(res.Name())
		}

	})

	if err := os.RemoveAll(ResultsDirPath()); err != nil {
		log.Printf("Could not delete %s with error %s", ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	if err := os.MkdirAll(ResultsDirPath(), 0777); err != nil {
		log.Printf("Could not create directory %s with error %s", ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	var err error
	if Observer, err = fsobserver.NewObserver(ResultsDirPath()); err != nil {
		log.Printf("Could not create directory %s observer with error %s", ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	StartNamesGenerator()
	log.SetFlags(log.Ltime)

	r.Run(fmt.Sprintf(":%d", Port))
}

// https://golang.org/src/net/http/status.go -- statuses
