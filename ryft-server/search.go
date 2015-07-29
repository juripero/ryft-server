package main

import (
	"log"
	"net/http"
	"os"

	"github.com/devicehive/ryft/rol"
)

func RawSearchProgress(s Search, n Names, adding, searching chan error) {
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
	resultsDs := func() {
		if s.Fuzziness == 0 {
			return ds.SearchExact(PathInRyftoneForResultDir(n.ResultFile), s.Query, s.Surrounding, "", &idxFile)
		} else {
			return ds.SearchFuzzy(PathInRyftoneForResultDir(n.ResultFile), s.Query, s.Surrounding, s.Fuzziness, "", &idxFile)
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
	addingFilesErr := <-addingFilesErrChan
	if addingFilesErr != nil {
		panic(addingFilesErr)
	}
}

func WaitingForSearchResults() (idxFile, resFile *os.File) {
	log.Println("Waiting for search result (and index)")
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
				}
			default:
			}
		}

		var err error

		if idxFile == nil {
			if idxFile, err = os.Open(ResultsDirPath(names.IdxFile)); err != nil {
				if os.IsNotExist(err) {
					continue
				}
				panic(&ServerError{http.StatusInternalServerError, err.Error()})
			}
			log.Printf("Index %s has been opened.", ResultsDirPath(names.IdxFile))
		}

		if resFile == nil {
			if resFile, err = os.Open(ResultsDirPath(names.ResultFile)); err != nil {
				if os.IsNotExist(err) {
					continue
				}
				panic(&ServerError{http.StatusInternalServerError, err.Error()})
			}
			log.Printf("Results %s has been opened.", ResultsDirPath(names.ResultFile))
		}

		log.Println("All files have been complete opened.")

		break
	}
}
