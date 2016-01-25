package main

import (
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/getryft/ryft-server/search"
	_ "github.com/getryft/ryft-server/search/ryftprim"
)

// just print of available search engines
func test0() {
	names := search.GetAvailableEngines()
	log.WithField("names", names).Infof("available search engines")

	for _, name := range names {
		engine, err := search.NewEngine(name, nil)
		if err != nil {
			log.WithField("name", name).WithError(err).
				Fatalf("failed to create search engine")
		}

		log.WithField("name", name).Infof("actual options: %+v", engine.Options())
	}
}

// run search locally
func test1a(tag string, engine search.Engine, cfg search.Config) {
	xlog := log.WithField("tag", tag)
	xlog.Printf("start /search: %s", cfg)

	res, err := engine.Search(&cfg)
	if err != nil {
		xlog.WithError(err).Fatalf("failed to start search")
	}

	var recordsReceived []*search.Record
	for {
		select {
		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				xlog.WithError(err).Errorf("search error received")
			}

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				xlog.WithField("data", rec).Debugf("new record received")
				recordsReceived = append(recordsReceived, rec)
			}

		case <-res.DoneChan:
			xlog.WithField("stat", res.Stat).Infof("/search finished")
			if uint64(len(recordsReceived)) != res.Stat.Matches {
				xlog.Warnf("%d matched but %d received",
					res.Stat.Matches, len(recordsReceived))
			}
			return // stop
		}
	}
}

// run multiple search: step-by-step
func test1b(engine search.Engine, cfgs ...search.Config) {
	for i, cfg := range cfgs {
		test1a(fmt.Sprintf("S%d", i+1), engine, cfg)
	}
}

// run multiple search: concurent
func test1c(engine search.Engine, cfgs ...search.Config) {
	wg := sync.WaitGroup{}

	// run each search in goroutine
	for i, cfg := range cfgs {
		wg.Add(1)
		cfg.Surrounding = uint(i + 1)
		go func(i int, cfg search.Config) {
			defer wg.Done()
			test1a(fmt.Sprintf("X%d", i+1), engine, cfg)
		}(i, cfg)

		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait() // wait all goroutines to finish
}

func test1(concurent bool) {
	backend := "ryftprim"
	opts := map[string]interface{}{
		"instance-name": "server-test",
		"keep-files":    true,
		"open-poll":     "1s",
		"read-poll":     "1s",
	}
	engine, err := search.NewEngine(backend, opts)
	if err != nil {
		log.WithError(err).Fatalf("failed to get search engine")
	}
	log.WithField("name", backend).Infof("actual options: %+v", engine.Options())

	A := search.NewConfig(`(RAW_TEXT CONTAINS "10")`, "/regression/*.txt")
	B := search.NewConfig(`(RAW_TEXT CONTAINS "0")`, "/regression/*.txt")
	C := search.NewConfig(`(RAW_TEXT CONTAINS "555")`, "/regression/*.txt")
	C.Fuzziness = 1

	cfgs := []search.Config{}
	cfgs = append(cfgs, *A, *B, *C)
	cfgs = append(cfgs, *A, *B, *C)
	cfgs = append(cfgs, *A, *B, *C)
	cfgs = append(cfgs, *A, *B, *C)
	cfgs = append(cfgs, *A, *B, *C)

	if !concurent {
		test1b(engine, cfgs...) // one-by-one
	} else {
		test1c(engine, cfgs...) // concurent
	}
}

func main() {
	//test0()
	//test1(false)
	test1(true) // concurent
}
