package main

import (
	"net/http"
	"os"

	"github.com/DataArt/ryft-rest-api/rol"
	"github.com/go-fsnotify/fsnotify"
)

func progress(s *Search, n Names, ch chan error) {
	ds := rol.RolDSCreate()
	defer ds.Delete()

	for _, f := range s.ExtractedFiles {
		ok := ds.AddFile(f)
		if !ok {
			ch <- &ServerError{http.StatusNotFound, "Could not add file " + f}
			return
		}
	}

	idxFile := PathInRyftoneForResultDir(n.IdxFile)
	resultsDs := func() *rol.RolDS {
		if s.Fuzziness == 0 {
			return ds.SearchExact(PathInRyftoneForResultDir(n.ResultFile), s.Query, s.Surrounding, "", &idxFile)
		} else {
			return ds.SearchFuzzyHamming(PathInRyftoneForResultDir(n.ResultFile), s.Query, s.Surrounding, s.Fuzziness, "", &idxFile)
		}
	}()
	defer resultsDs.Delete()

	if err := resultsDs.HasErrorOccured(); err != nil {
		if !err.IsStrangeError() {
			ch <- &ServerError{http.StatusInternalServerError, err.Error()}
			return
		}
	}

	ch <- nil
}

func startAndWaitFiles(s *Search, n Names, ch chan error) (idxFile, resFile *os.File, idxops, resops chan fsnotify.Op) {
	var err error

	idxPath := ResultsDirPath(n.IdxFile)
	resPath := ResultsDirPath(n.ResultFile)

	idxops = Observer.Follow(idxPath, 16)
	resops = Observer.Follow(resPath, 16)

	go progress(s, n, ch)

	var idxCreated, resCreated bool

ops:
	for !(idxCreated && resCreated) {
		select {
		case idxop := <-idxops:
			if !idxCreated && idxop&fsnotify.Create == fsnotify.Create {
				idxCreated = true
			}
		case resop := <-resops:
			if !resCreated && resop&fsnotify.Create == fsnotify.Create {
				resCreated = true
			}
		case err = <-ch:
			if err != nil {
				panic(err)
			} else {
				break ops
			}
		}
	}

	if idxFile, err = os.Open(idxPath); err != nil {
		panic(&ServerError{http.StatusInternalServerError, err.Error()})
	}

	if resFile, err = os.Open(resPath); err != nil {
		panic(&ServerError{http.StatusInternalServerError, err.Error()})
	}
	return

}
