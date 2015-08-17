package main

import (
	"net/http"
	"os"

	"github.com/DataArt/ryft-rest-api/rol"
	"github.com/go-fsnotify/fsnotify"
)

// func RawSearchProgress(s *Search, n Names, adding, searching chan error) {
// ds := rol.RolDSCreate()
// defer ds.Delete()

// for _, f := range s.ExtractedFiles {
// 	ok := ds.AddFile(f)
// 	if !ok {
// 		adding <- &ServerError{http.StatusNotFound, "Could not add file " + f}
// 	}
// }
// adding <- nil

// idxFile := PathInRyftoneForResultDir(n.IdxFile)
// resultsDs := func() *rol.RolDS {
// 	if s.Fuzziness == 0 {
// 		return ds.SearchExact(PathInRyftoneForResultDir(n.ResultFile), s.Query, s.Surrounding, "", &idxFile)
// 	} else {
// 		return ds.SearchFuzzyHamming(PathInRyftoneForResultDir(n.ResultFile), s.Query, s.Surrounding, s.Fuzziness, "", &idxFile)
// 	}
// }()
// defer resultsDs.Delete()

// if err := resultsDs.HasErrorOccured(); err != nil {
// 	if !err.IsStrangeError() {
// 		searching <- &ServerError{http.StatusInternalServerError, err.Error()}
// 	}
// }

// searching <- nil
// }

// func ProcessAddingFilesError(adding chan error) {
// 	if e := <-adding; e != nil {
// 		panic(e)
// 	}
// }

func progress(s *Search, n Names, ch chan error) {
	ds := rol.RolDSCreate()
	defer ds.Delete()

	for _, f := range s.ExtractedFiles {
		ok := ds.AddFile(f)
		if !ok {
			ch <- &ServerError{http.StatusNotFound, "Could not add file " + f}
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
			searching <- &ServerError{http.StatusInternalServerError, err.Error()}
		}
	}

	ch <- nil
}

func startAndWaitFiles(s *Search, n Names, ch chan error) (idxFile, resFile *os.File, idxops, resops chan fsnotify.Op) {
	var err error

	idxPath := ResultsDirPath(n.IdxFile)
	resPath := ResultsDirPath(n.ResultFile)

	idxops = Observer.Follow(idxPath)
	resops = Observer.Follow(resPath)

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

// func startAndWait(s *Search, n Names, ch chan error) (idxFile, resFile *os.File, w *fsnotify.Watcher) {
// 	var err error

// 	if w, err = fsnotify.NewWatcher(); err != nil {
// 		panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 	}

// 	if err = w.Add(ResultsDirPath()); err != nil {
// 		panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 	}

// idxPath := ResultsDirPath(n.IdxFile)
// resPath := ResultsDirPath(n.ResultFile)

// 	go progress(s, n, ch)

// 	var idxCreated, resCreated bool
// 	for !(idxCreated && resCreated) {
// 		select {
// 		case e := <-w.Events:
// 			if e.Op&fsnotify.Create == fsnotify.Create {
// 				if !idxCreated {
// 					if e.Name == idxPath {
// 						idxCreated == true
// 						continue
// 					}
// 				}

// 				if !resCreated {
// 					if e.Name == resPath {
// 						resCreated == true
// 						continue
// 					}

// 				}
// 			}
// 		case err = <-w.Errors:
// 			log.Printf("Watcher error: %s")
// 		case err = <-ch:
// 			panic(err)
// 		}
// 	}

// if idxFile, err = os.Open(idxPath); err != nil {
// 	panic(&ServerError{http.StatusInternalServerError, err.Error()})
// }

// if resFile, err = os.Open(resPath); err != nil {
// 	panic(&ServerError{http.StatusInternalServerError, err.Error()})
// }

// return
// }

// func WaitingForSearchResults(n Names, searching chan error, sleepiness time.Duration) (idxFile, resFile *os.File) {
// 	log.Println("Waiting for search result (and index)")
// 	var searchErr error = nil
// 	var searchErrReady bool = false
// 	for {
// 		if !searchErrReady {
// 			select {
// 			case searchErr = <-searching:
// 				searchErrReady = true
// 				if searchErr != nil {
// 					log.Printf("Error in search channel: %s", searchErr)
// 					panic(searchErr)
// 				}
// 			default:
// 			}
// 		}

// 		var err error

// 		if idxFile == nil {
// 			if idxFile, err = os.Open(ResultsDirPath(n.IdxFile)); err != nil {
// 				if os.IsNotExist(err) {
// 					time.Sleep(sleepiness)
// 					continue
// 				}
// 				panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 			}
// 			log.Printf("Index %s has been opened.", ResultsDirPath(n.IdxFile))
// 		}

// 		if resFile == nil {
// 			if resFile, err = os.Open(ResultsDirPath(n.ResultFile)); err != nil {
// 				if os.IsNotExist(err) {
// 					time.Sleep(sleepiness)
// 					continue
// 				}
// 				panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 			}
// 			log.Printf("Results %s has been opened.", ResultsDirPath(n.ResultFile))
// 		}

// 		log.Println("All files have been complete opened.")

// 		break
// 	}
// 	return
// }

// func WaitingForResultsOld(n Names, s chan error) (idxFile, resFile *os.File) {
// 	var idxw, resw *fsnotify.Watcher
// 	var err error

// 	idxPath := ResultsDirPath(n.IdxFile)
// 	resPath := ResultsDirPath(n.ResultFile)

// 	if idxw, err = fsnotify.NewWatcher(); err != nil {
// 		panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 	}
// 	defer idxw.Close()

// 	if err = idxw.Add(idxPath); err != nil {
// 		panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 	}

// 	if resw, err = fsnotify.NewWatcher(); err != nil {
// 		panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 	}
// 	defer resw.Close()

// 	if err = resw.Add(resPath); err != nil {
// 		panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 	}

// 	log.Printf("Created watchers for %s & %s", idxPath, resPath)
// 	var idxCreated, resCreated bool
// 	for !(idxCreated && resCreated) {
// 		select {
// 		case idxEvent := <-idxw.Events:
// 			log.Printf("Received idxEvent=%d", idxEvent)
// 			if !idxCreated {
// 				if idxEvent.Op&fsnotify.Create == fsnotify.Create {
// 					log.Println("idxEvent contains creation flag")
// 					idxCreated = true
// 					idxw.Close()
// 				}
// 			}
// 		case idxErr := <-idxw.Errors:
// 			panic(&ServerError{http.StatusInternalServerError, idxErr.Error()})
// 		case resEvent := <-resw.Events:
// 			log.Printf("Received resEvent=%d", resEvent)
// 			if !resCreated {
// 				if resEvent.Op&fsnotify.Create == fsnotify.Create {
// 					log.Println("resEvent contains creation flag")
// 					resCreated = true
// 					resw.Close()
// 				}
// 			}
// 		case resErr := <-resw.Errors:
// 			panic(&ServerError{http.StatusInternalServerError, resErr.Error()})
// 		case err = <-s:
// 			if err != nil {
// 				panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 			}
// 		}
// 	}

// 	if idxFile, err = os.Open(idxPath); err != nil {
// 		panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 	}

// 	if resFile, err = os.Open(resPath); err != nil {
// 		panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 	}

// 	return
// }

// func WaitingForResults(n Names, s chan error) (idxFile, resFile *os.File) {
// 	var w *fsnotify.Watcher
// 	var err error
// 	if w, err = fsnotify.NewWatcher(); err != nil {
// 		panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 	}

// 	if err = w.Add(ResultsDirPath()); err != nil {
// 		panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 	}

// 	var idxCreated, resCreated bool
// 	// trying open files

// 	idxPath := ResultsDirPath(n.IdxFile)
// 	resPath := ResultsDirPath(n.ResultFile)

// 	for !(idxCreated && resCreated) {
// 		select {
// 		case e := <-w.Events:
// 			if e.Op&fsnotify.Create == fsnotify.Create {
// 				if !idxCreated {
// 					if e.Name == idxPath {
// 						idxCreated == true
// 						continue
// 					}
// 				}

// 				if !resCreated {
// 					if e.Name == resPath {
// 						resCreated == true
// 						continue
// 					}

// 				}
// 			}
// 		case err = <-w.Errors:
// 			log.Printf("Watcher error: %s")
// 		}
// 	}
// 	w.Close()

// }
