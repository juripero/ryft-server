package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/devicehive/ryft/rol"
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

func main() {
	r := gin.Default()

	r.GET("/search/test-ok", func(c *gin.Context) {
		defer deferRecover(c)
		c.Writer.Header()["Content-Type"] = []string{binding.MIMEJSON}
		c.Stream(func(w io.Writer) bool {
			w.Write([]byte("["))
			for i := 0; i <= 100; i++ {

				record := gin.H{"number": i}
				bytes, err := json.Marshal(record)
				if err != nil {
					panic(&ServerError{http.StatusInternalServerError, err.Error()})
				}

				w.Write(bytes)
				w.Write(",")
			}

			w.Write([]byte("]"))
			return true
		})

	})

	r.GET("/search/test-fail", func(c *gin.Context) {
		defer deferRecover(c)
		panic(&ServerError{http.StatusInternalServerError, fmt.Errorf("Test error")})
	})

	r.GET("/search/exact", func(c *gin.Context) {
		defer deferRecover(c)
		s := new(Search)
		if err := c.Bind(s); err != nil {
			panic(&ServerError{http.StatusBadRequest, err.Error()})
		}

		s.ExtractFiles()

		names := GetNewNames()

		addingFiles := make(chan error)
		searchProblem := make(chan error)
		go func() {
			ds := rol.RolDSCreate()
			defer ds.Delete()

			for _, f := range s.ExtractedFiles {
				ok := ds.AddFile(f)
				if !ok {
					addingFiles <- &ServerError(http.StatusNotFound, fmt.Sprintf("Could not add file `%s`", f))
				}
			}
			addingFiles <- nil

			idxFile := PathInRyftoneForResultDir(names.IdxFile)
			resultsDs := ds.SearchExact(PathInRyftoneForResultDir(names.ResultFile), s.Query, s.Surrounding, "", &idx)

		}()

		addingFilesErr := <-addingFiles
		if addingFilesErr != nil {
			panic(addingFilesErr)
		}

		var idxFile, resFile os.File
		var openErr error

	waitingForResults:
		for {
			select {
			case openErr := <-searchProblem:
				panic(openErr)
			default:
				var err error
				if idxFile, err = os.Open(PathInRyftoneForResultDir(names.IdxFile)); err != nil {
					if os.IsNotExist(err) {
						continue
					}
					panic(&ServerError(http.StatusInternalServerError, err.Error()))
				}

				if resFile, err = os.Open(PathInRyftoneForResultDir(names.ResultFile)); err != nil {
					if os.IsNotExist(err) {
						continue
					}
					panic(&ServerError(http.StatusInternalServerError, err.Error()))
				}
				break waitingForResults
			}
		}

	})

	// remove directory Res directory
	// create a direcotry ResDirectory

	StartNamesGenerator()

	r.Run(":8765")
}

/* Help
https://golang.org/src/net/http/status.go -- statuses
*/

/* Ready fro requests
http://localhost:8765/search/exact?query=%28%20RAW_TEXT%20CONTAINS%20%22night%22%20%29&files=passengers.txt&surrounding=10

*/
