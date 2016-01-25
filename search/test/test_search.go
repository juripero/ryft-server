package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/getryft/ryft-server/search"
	_ "github.com/getryft/ryft-server/search/ryftprim"
)

// just print of available search engines
func test0() {
	names := search.GetAvailableEngines()
	log.Printf("engines: %s", names)

	for _, name := range names {
		engine, err := search.NewEngine(name, nil)
		if err != nil {
			log.Fatalf("%q: failed to create search engine: %s", name, err)
		}

		log.Printf("%q options: %+v", name, engine.Options())
	}
}

// run search locally
func test1a(id string, engine search.Engine, cfg search.Config) {
	log.Printf("%q search starting...: %s", id, cfg)
	res := search.NewResult()
	err := engine.Search(&cfg, res)
	if err != nil {
		log.Fatalf("failed to start %q search: %s", id, err)
	}

	var recordsReceived []*search.Record
	for {
		select {
		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				log.Printf("%q search error: %s", id, err)
			}

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				_ = rec // log.Printf("%q result: %s", id, rec)
				recordsReceived = append(recordsReceived, rec)
			}

		case <-res.DoneChan:
			// log.Printf("%q search done: no more records", id)
			log.Printf("%q search statistics: %s", id, res.Stat)
			if uint64(len(recordsReceived)) != res.Stat.Matches {
				log.Printf("%q WARN: %d matched but %d received",
					id, res.Stat.Matches, len(recordsReceived))
				for _, r := range recordsReceived {
					log.Printf("%q: %s", id, r)
				}
			}
			return
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
		cfg.Surrounding = uint16(i + 1)
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
		log.Fatalf("failed to get search engine: %s", err)
	}
	log.Printf("%q actual options: %+v", backend, engine.Options())

	A := search.NewConfig(`(RAW_TEXT CONTAINS "10")`)
	A.AddFiles("/regression/*.txt")
	A.Surrounding = 0
	A.Fuzziness = 0

	B := search.NewConfig(`(RAW_TEXT CONTAINS "0")`)
	B.AddFiles("/regression/*.txt")
	B.Surrounding = 0
	B.Fuzziness = 0

	C := search.NewConfig(`(RAW_TEXT CONTAINS "555")`)
	C.AddFiles("/regression/*.txt")
	C.Surrounding = 0
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
