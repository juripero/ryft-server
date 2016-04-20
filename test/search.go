package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftdec"
	"github.com/getryft/ryft-server/search/ryftmux"
)

var (
	// ryfthttp
	ryfthttpServerUrl = "http://localhost:8765"
	ryfthttpLogLevel  = "warn"

	// ryftprim
	ryftprimInstance = ".test"
	ryftprimLogLevel = "warn"

	// ryftone
	ryftoneInstance = ".test"
	ryftoneLogLevel = "warn"

	// ryftmux
	ryftmuxLogLevel = "warn"

	// ryftdec
	ryftdecLogLevel = "warn"

	printReceivedRecords = false
)

// just print of available search engines
// also print default options for each engine
func printSearchEngines(log Logger) {
	names := search.GetAvailableEngines()
	log("available search engines: %q", names)

	for _, name := range names {
		engine, err := search.NewEngine(name, nil)
		if err != nil {
			log("failed to create %q search engine: %s", name, err)
			panic(err)
		}

		log("%q: default options: %+v", name, engine.Options())
	}
}

// create new search engine
func newEngine(log Logger, backend string, opts map[string]interface{}) search.Engine {
	engine, err := search.NewEngine(backend, opts)
	if err != nil {
		log("failed to get %q search engine: %s", backend, err)
		panic(err)
	}
	log("%q: actual options: %+v", backend, engine.Options())

	return engine
}

// create new ryftprim search engine
func newRyftPrim(log Logger) search.Engine {
	opts := map[string]interface{}{
		"instance-name": ryftprimInstance,
		"keep-files":    true,
		"open-poll":     "1s",
		"read-poll":     "1s",
		"read-limit":    5,
		"log-level":     ryftprimLogLevel,
	}

	return newEngine(log, "ryftprim", opts)
}

// create new ryftone search engine
func newRyftOne(log Logger) search.Engine {
	opts := map[string]interface{}{
		"instance-name": ryftoneInstance,
		"keep-files":    true,
		"open-poll":     "1s",
		"read-poll":     "1s",
		"read-limit":    5,
		"log-level":     ryftoneLogLevel,
	}

	return newEngine(log, "ryftone", opts)
}

// create new ryfthttp search engine
func newRyftHttp(log Logger) search.Engine {
	opts := map[string]interface{}{
		"server-url": ryfthttpServerUrl,
		"local-only": true,
		"log-level":  ryfthttpLogLevel,
	}

	return newEngine(log, "ryfthttp", opts)
}

// create new ryftdec search engine
func newRyftDec(log Logger, backend search.Engine) search.Engine {
	name := "ryftdec"
	engine, err := ryftdec.NewEngine(backend)
	if err != nil {
		log("failed to get %q search engine: %s", name, err)
		panic(err)
	}
	// log("%q: actual options: %+v", name, engine.Options())

	ryftdec.SetLogLevel(ryftdecLogLevel)
	return engine
}

// create new ryftmux search engine
func newRyftMux(log Logger, backends ...search.Engine) search.Engine {
	name := "ryftmux"
	engine, err := ryftmux.NewEngine(backends...)
	if err != nil {
		log("failed to get %q search engine: %s", name, err)
		panic(err)
	}
	// log("%q: actual options: %+v", name, engine.Options())

	ryftmux.SetLogLevel(ryftmuxLogLevel)
	return engine
}

// combined search results
type SearchResult struct {
	Errors  []error
	Records []*search.Record
	Stat    *search.Statistics
}

// collect all results from the search engine
func grabResults(log Logger, tag string, res *search.Result, checkRecords bool) (r SearchResult) {
	start := time.Now()
	defer func() {
		stop := time.Now()
		log("[%s]: processing time: %s", tag, stop.Sub(start))
	}()

	// preallocate buffer
	r.Records = make([]*search.Record, 0, 16*1024)

	for {
		select {
		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				log("[%s]: error received: %s", tag, err)
				r.Errors = append(r.Errors, err)
				continue
			}

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				if printReceivedRecords {
					log("[%s]: record received: %s", tag, rec)
				}
				r.Records = append(r.Records, rec)
				continue
			}

		case <-res.DoneChan:
			if res.Stat != nil {
				log("[%s]: finished with %s", tag, res.Stat)
				r.Stat = res.Stat
			} else {
				log("[%s]: finished with no stat", tag)
			}

			// drain error channel
			for err := range res.ErrorChan {
				log("[%s]: *** error received: %s", tag, err)
				r.Errors = append(r.Errors, err)
			}

			// drain record channel
			for rec := range res.RecordChan {
				if printReceivedRecords {
					log("[%s]: *** record received: %s", tag, rec)
				}
				r.Records = append(r.Records, rec)
			}

			if checkRecords && res.Stat != nil && uint64(len(r.Records)) != res.Stat.Matches {
				log("[%s] WARNING: %d matched but %d received",
					tag, res.Stat.Matches, len(r.Records))
			}

			return // stop
		}
	}
}

// run one search
func runSearch1(log Logger, tag string, engine search.Engine, cfg *search.Config) SearchResult {
	log("[%s]: start /search: %s", tag, cfg)

	res, err := engine.Search(cfg)
	if err != nil {
		log("[%s]: failed to start /search: %s", tag, err)
		panic(err)
	}

	return grabResults(log, tag, res, true)
}

// run multiple search: step-by-step
func runSearchOneByOne(log Logger, engine search.Engine, cfgs ...search.Config) {
	for i, cfg := range cfgs {
		tag := fmt.Sprintf("S%d", i+1)
		runSearch1(log, tag, engine, &cfg)
	}
}

// run multiple search: concurent
func runSearchConcurent(log Logger, engine search.Engine, cfgs ...search.Config) {
	wg := sync.WaitGroup{}

	// run each search in goroutine
	for i, cfg := range cfgs {
		wg.Add(1)
		go func(i int, cfg search.Config) {
			defer wg.Done()
			tag := fmt.Sprintf("X%d", i+1)
			runSearch1(log, tag, engine, &cfg)
		}(i, cfg)

		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait() // wait all goroutines to finish
}
