package main

import (
	"encoding/json"
	"io"
	"log"
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

	r.GET("/search/exact", func(c *gin.Context) {
		defer deferRecover(c)
		s := new(Search)
		if err := c.Bind(s); err != nil {
			panic(&ServerError{http.StatusBadRequest, err.Error()})
		}

		s.ExtractFiles()

		names := GetNewNames()

		addingFilesErrChan := make(chan error)
		searchingErrChan := make(chan error)
		go func() {
			ds := rol.RolDSCreate()
			defer ds.Delete()

			for _, f := range s.ExtractedFiles {
				ok := ds.AddFile(f)
				if !ok {
					addingFilesErrChan <- &ServerError{http.StatusNotFound, "Could not add file " + f}
				}
			}
			addingFilesErrChan <- nil

			idxFile := PathInRyftoneForResultDir(names.IdxFile)
			resultsDs := ds.SearchExact(PathInRyftoneForResultDir(names.ResultFile), s.Query, s.Surrounding, "", &idxFile)
			defer resultsDs.Delete()

			if err := resultsDs.HasErrorOccured(); err != nil {
				if !err.IsStrangeError() {
					searchingErrChan <- &ServerError{http.StatusInternalServerError, err.Error()}
				}
			}

			searchingErrChan <- nil

		}()

		addingFilesErr := <-addingFilesErrChan
		if addingFilesErr != nil {
			panic(addingFilesErr)
		}

		log.Println("Start search listening")

		var idxFile, resFile *os.File
		var searchErr error = nil
		var searchErrReady bool = false

		for {
			if !searchErrReady {
				select {
				case searchErr = <-searchingErrChan:
					searchErrReady = true
					if searchErr != nil {
						log.Printf("Error in search channel: %s", searchErr)
						panic(searchErr)
					} else {
						log.Println("SearchErr==nil, but files is not created for this time.")
					}

				default:
					log.Println("No information about searching status. Continue...")
				}
			}

			var err error
			if idxFile, err = os.Open(ResultsDirPath(names.IdxFile)); err != nil {
				if os.IsNotExist(err) {
					log.Printf("Index %s do not exists. Continue...", ResultsDirPath(names.IdxFile))
					continue
				}
				panic(&ServerError{http.StatusInternalServerError, err.Error()})
			}
			log.Printf("Index %s has been opened.", ResultsDirPath(names.IdxFile))

			if resFile, err = os.Open(ResultsDirPath(names.ResultFile)); err != nil {
				if os.IsNotExist(err) {
					log.Printf("Results %s do not exists. Continue...", ResultsDirPath(names.ResultFile))
					continue
				}
				panic(&ServerError{http.StatusInternalServerError, err.Error()})
			}
			log.Println("Results %s has been opened.", ResultsDirPath(names.ResultFile))
			break
		}

		idxFile.Close()
		resFile.Close()
		c.IndentedJSON(http.StatusOK, gin.H{"completion": "ok"})
		log.Println("Processing request complete")

		// 	c.Stream(func(w io.Writer) bool {
		// 		w.Write([]byte("["))
		// 		StreamJsonContentOfArray(resFile, idxFile, w, false)
		// 		w.Write([]byte("]"))
		// 		return false
		// 	})
	})

	if err := os.RemoveAll(ResultsDirPath()); err != nil {
		log.Printf("Could not delete %s with error %s", ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	if err := os.MkdirAll(ResultsDirPath(), 0777); err != nil {
		log.Printf("Could not create directory %s with error %s", ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	StartNamesGenerator()

	r.Run(":8765")
}

/* Help
https://golang.org/src/net/http/status.go -- statuses
*/

/* Ready fro requests
http://localhost:8765/search/exact?query=%28%20RAW_TEXT%20CONTAINS%20%22night%22%20%29&files=passengers.txt&surrounding=10
http://192.168.56.103:8765/search/exact?query=( RAW_TEXT CONTAINS "night" )&files=passengers.txt&surrounding=10

*/
