package main

import (
	"log"
	"net/http"
	"os"

	"github.com/DataArt/ryft-rest-api/rol"
	"github.com/go-fsnotify/fsnotify"
)

func progress(s *Search, n Names, ch chan error) {
	log.Println("progress: start")
	ds := rol.RolDSCreate()
	defer ds.Delete()

	for _, f := range s.ExtractedFiles {
		ok := ds.AddFile(f)
		if !ok {
			ch <- &ServerError{http.StatusNotFound, "Could not add file " + f}
			return
		}
		log.Printf("progress: added file %s", f)
	}

	idxFile := PathInRyftoneForResultDir(n.IdxFile)
	resultsDs := func() *rol.RolDS {
		if s.Fuzziness == 0 {
			log.Println("progress: starting exact search")
			return ds.SearchExact(PathInRyftoneForResultDir(n.ResultFile), s.Query, s.Surrounding, "", &idxFile)
		} else {
			log.Println("progress: starting fuzzy-hamming search")
			return ds.SearchFuzzyHamming(PathInRyftoneForResultDir(n.ResultFile), s.Query, s.Surrounding, s.Fuzziness, "", &idxFile)
		}
	}()
	log.Println("progress: API-search completed")
	defer resultsDs.Delete()

	if err := resultsDs.HasErrorOccured(); err != nil {
		if !err.IsStrangeError() {
			log.Printf("progress: end; API-search completed with error: %s", err.Error())
			searching <- &ServerError{http.StatusInternalServerError, err.Error()}
			return
		}
	}

	ch <- nil
	log.Println("progress: end")
}

func startAndWaitFiles(s *Search, n Names, ch chan error) (idxFile, resFile *os.File, idxops, resops chan fsnotify.Op) {
	log.Println("waiting: start")
	var err error

	idxPath := ResultsDirPath(n.IdxFile)
	resPath := ResultsDirPath(n.ResultFile)

	idxops = Observer.Follow(idxPath)
	resops = Observer.Follow(resPath)
	log.Printf("waiting: followed %s and %s", idxPath, resPath)

	go progress(s, n, ch)

	var idxCreated, resCreated bool

ops:
	for !(idxCreated && resCreated) {
		log.Printf("waiting: events iteration: idxCreated=%+v resCreated=%+v", idxCreated, resCreated)
		select {
		case idxop := <-idxops:
			log.Printf("waiting: select: receive idxop=%+v", idxop)
			if !idxCreated && idxop&fsnotify.Create == fsnotify.Create {
				log.Printf("waiting: select: receive idxop=%+v -> create", idxop)
				idxCreated = true
			}
		case resop := <-resops:
			log.Printf("waiting: select: receive resop=%+v", resop)
			if !resCreated && resop&fsnotify.Create == fsnotify.Create {
				log.Printf("waiting: select: receive resop=%+v -> create", resop)
				resCreated = true
			}
		case err = <-ch:
			log.Printf("waiting: select: receive ch=%s", err.Error())
			if err != nil {
				panic(err)
			} else {
				break ops
			}
		}
	}

	log.Println("waiting: iterations complete")

	if idxFile, err = os.Open(idxPath); err != nil {
		panic(&ServerError{http.StatusInternalServerError, err.Error()})
	}
	log.Printf("waiting: idx opened: %s", idxPath)

	if resFile, err = os.Open(resPath); err != nil {
		panic(&ServerError{http.StatusInternalServerError, err.Error()})
	}
	log.Printf("waiting: res opened: %s", resPath)
	return

}
