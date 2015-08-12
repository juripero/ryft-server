package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DataArt/ryft-rest-api/rol"
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
