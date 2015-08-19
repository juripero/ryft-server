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
	"time"

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
		ch := make(chan error, 1)

		log.Printf("request: start waiting for files %+v", n)
		idx, res, idxops, resops := startAndWaitFiles(s, n, ch)
		defer func() {
			Observer.Unfollow(idx.Name())
			Observer.Unfollow(res.Name())
			idx.Close()
			res.Close()
			log.Println("request: ops & files closed")
		}()
		log.Println("request: all files created & opened")

		dropper := make(chan struct{}, 1)
		records := GetRecordsChan(idx, idxops, ch, dropper)

		c.Stream(func(w io.Writer) bool {
			generateJson(records, res, resops, w, dropper)
			log.Println("request: after generateJson")
			return false
		})
		log.Println("request: after stream loop")

		if !KeepResults {
			os.Remove(idx.Name())
			os.Remove(res.Name())
			log.Println("request: file deleted")
		}

		log.Println("request: end")

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

	//r.Run(fmt.Sprintf(":%d", Port))
	CustomRun(r, fmt.Sprintf(":%d", Port))
}

// func (engine *Engine) Run(addr ...string) (err error) {
// defer func() { debugPrintError(err) }()

// address := resolveAddress(addr)
// debugPrint("Listening and serving HTTP on %s\n", address)
// err = http.ListenAndServe(address, engine)
// return
// }

func CustomRun(engine *gin.Engine, addr ...string) (err error) {
	address := resolveAddress(addr)
	debugPrint("Listening and serving HTTP on %s\n", address)
	err = CustomListenAndServe(address, engine)
	return
}

func CustomListenAndServe(addr string, handler http.Handler) error {
	server := &http.Server{Addr: addr, Handler: handler}
	server.WriteTimeout = 60 * time.Second
	return server.ListenAndServe()
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		if port := os.Getenv("PORT"); len(port) > 0 {
			debugPrint("Environment variable PORT=\"%s\"", port)
			return ":" + port
		} else {
			debugPrint("Environment variable PORT is undefined. Using port :8080 by default")
			return ":8080"
		}
	case 1:
		return addr[0]
	default:
		panic("too much parameters")
	}
}

func debugPrint(format string, values ...interface{}) {
	if gin.IsDebugging() {
		log.Printf("[GIN-debug] "+format, values...)
	}
}

// https://golang.org/src/net/http/status.go -- statuses
