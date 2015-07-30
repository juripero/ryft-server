package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

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

func rawSearchHandler(isFuzzy bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		defer deferRecover(c)
		s := new(Search)
		if err := c.Bind(s); err != nil {
			panic(&ServerError{http.StatusBadRequest, err.Error()})
		}

		if !isFuzzy {
			s.Fuzziness = 0
		}

		s.ExtractFiles()

		names := GetNewNames()

		addingFilesErrChan := make(chan error)
		searchingErrChan := make(chan error)
		go RawSearchProgress(s, names, addingFilesErrChan, searchingErrChan)

		ProcessAddingFilesError(addingFilesErrChan)

		resFile, idxFile := WaitingForSearchResults(names, searchingErrChan)

		c.Stream(func(w io.Writer) bool {
			StreamJson(resFile, idxFile, w, searchingErrChan)
			return false
		})

		idxFile.Close()
		resFile.Close()

		log.Println("Processing request complete")
	}
}

func main() {
	r := gin.Default()

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

	r.GET("/search/exact", rawSearchHandler(false))
	r.GET("/search/fuzzy", rawSearchHandler(true))

	if err := os.RemoveAll(ResultsDirPath()); err != nil {
		log.Printf("Could not delete %s with error %s", ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	if err := os.MkdirAll(ResultsDirPath(), 0777); err != nil {
		log.Printf("Could not create directory %s with error %s", ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	StartNamesGenerator()
	log.SetFlags(log.Ltime)

	r.Run(":8765")
}

/* Help
https://golang.org/src/net/http/status.go -- statuses
*/

/* Ready fro requests
http://localhost:8765/search/exact?query=%28%20RAW_TEXT%20CONTAINS%20%22night%22%20%29&files=passengers.txt&surrounding=10
http://192.168.56.103:8765/search/exact?query=( RAW_TEXT CONTAINS "night" )&files=passengers.txt&surrounding=10

*/
