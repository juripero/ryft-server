package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DataArt/ryft-rest-api/rol"
	"github.com/go-fsnotify/fsnotify"
)

func RawSearchProgress(s *Search, n Names, adding, searching chan error) {
	ds := rol.RolDSCreate()
	defer ds.Delete()

	for _, f := range s.ExtractedFiles {
		ok := ds.AddFile(f)
		if !ok {
			adding <- &ServerError{http.StatusNotFound, "Could not add file " + f}
		}
	}
	adding <- nil

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
			searching <- &ServerError{http.StatusInternalServerError, err.Error()}
		}
	}

	searching <- nil
}

func ProcessAddingFilesError(adding chan error) {
	if e := <-adding; e != nil {
		panic(e)
	}
}

func WaitingForSearchResults(n Names, searching chan error, sleepiness time.Duration) (idxFile, resFile *os.File) {
	log.Println("Waiting for search result (and index)")
	var searchErr error = nil
	var searchErrReady bool = false
	for {
		if !searchErrReady {
			select {
			case searchErr = <-searching:
				searchErrReady = true
				if searchErr != nil {
					log.Printf("Error in search channel: %s", searchErr)
					panic(searchErr)
				}
			default:
			}
		}

		var err error

		if idxFile == nil {
			if idxFile, err = os.Open(ResultsDirPath(n.IdxFile)); err != nil {
				if os.IsNotExist(err) {
					time.Sleep(sleepiness)
					continue
				}
				panic(&ServerError{http.StatusInternalServerError, err.Error()})
			}
			log.Printf("Index %s has been opened.", ResultsDirPath(n.IdxFile))
		}

		if resFile == nil {
			if resFile, err = os.Open(ResultsDirPath(n.ResultFile)); err != nil {
				if os.IsNotExist(err) {
					time.Sleep(sleepiness)
					continue
				}
				panic(&ServerError{http.StatusInternalServerError, err.Error()})
			}
			log.Printf("Results %s has been opened.", ResultsDirPath(n.ResultFile))
		}

		log.Println("All files have been complete opened.")

		break
	}
	return
}

func WaitingForResults(n Names, s chan err) (idxFile, resFile *os.File) {
	var idxw, resw *fsnotify.Watcher
	var err error

	idxPath := ResultsDirPath(n.IdxFile)
	resPath := ResultsDirPath(n.ResultFile)

	if idxw, err = fsnotify.NewWatcher(); err != nil {
		panic(&ServerError{http.StatusInternalServerError, err.Error()})
	}
	defer idxw.Close()

	if err = idxw.Add(idxPath); err != nil {
		panic(&ServerError{http.StatusInternalServerError, err.Error()})
	}

	if resw, err = fsnotify.NewWatcher(); err != nil {
		panic(&ServerError{http.StatusInternalServerError, err.Error()})
	}
	defer resw.Close()

	if err = resw.Add(respPath); err != nil {
		panic(&ServerError{http.StatusInternalServerError, err.Error()})
	}

	log.Printf("Created watchers for %s & %s", idxPath, resPath)
	var idxCreated, resCreated bool
	for !(idxCreated && resCreated) {
		select {
		case idxEvent := <-idxw.Events:
			log.Printf("Received idxEvent=%d", idxEvent)
			if !idxCreated {
				if idxEvent.Op&fsnotify.Create == fsnotify.Create {
					log.Println("idxEvent contains creation flag")
					idxCreated = true
					idxw.Close()
				}
			}
		case idxErr := <-idxw.Errors:
			panic(&ServerError{http.StatusInternalServerError, idxErr.Error()})
		case resEvent := <-resw.Events:
			log.Printf("Received resEvent=%d", idxEvent)
			if !resCreated {
				if resEvent.Op&fsnotify.Create == fsnotify.Create {
					log.Println("resEvent contains creation flag")
					resCreated = true
					resw.Close()
				}
			}
		case resErr := <-resw.Errors:
			panic(&ServerError{http.StatusInternalServerError, resErr.Error()})
		case err = <-s:
			if err != nil {
				panic(&ServerError{http.StatusInternalServerError, err.Error()})
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
