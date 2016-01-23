package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/getryft/ryft-server/search"
	_ "github.com/getryft/ryft-server/search/ryftprim"
)

// just print of available search engines
func test0() {
	names := search.GetAvailableEngines()
	log.Printf("engines: %s", names)
}

// run search locally
func test1a(id string, engine search.Engine, cfg *search.Config) {
	log.Printf("%q search starting...: %s", id, cfg)
	res := search.NewResult()
	err := engine.Search(cfg, res)
	if err != nil {
		log.Fatalf("failed to start %q search: %s", id, err)
	}

	for {
		select {
		case err, ok := <-res.ErrorChan:
			if ok {
				log.Printf("%q search error: %s", id, err)
			}

		case rec, ok := <-res.RecordChan:
			if ok {
				log.Printf("%q result: %s", id, rec)
			} else {
				log.Printf("%q search done: no more records", id)
				return
			}
		}
	}
}

// run multiple search: step-by-step
func test1b(engine search.Engine, cfgs ...*search.Config) {
	for i, cfg := range cfgs {
		test1a(fmt.Sprintf("S%d", i+1), engine, cfg)
	}
}

// run multiple search: concurent
func test1c(engine search.Engine, cfgs ...*search.Config) {
	var wg sync.WaitGroup
	wg.Add(len(cfgs))

	for i, cfg := range cfgs {
		go func(i int, cfg *search.Config) {
			defer wg.Done()
			test1a(fmt.Sprintf("X%d", i+1), engine, cfg)
		}(i, cfg)
	}

	wg.Wait()
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

	A := search.NewConfig(`(RAW_TEXT CONTAINS "10")`)
	A.AddFiles("regression/*.txt")
	A.Surrounding = 10
	A.Fuzziness = 0

	B := search.NewConfig(`(RAW_TEXT CONTAINS "0")`)
	B.AddFiles("regression/*.txt")
	B.Surrounding = 8
	B.Fuzziness = 0

	C := search.NewConfig(`(RAW_TEXT CONTAINS "555")`)
	C.AddFiles("regression/*.txt")
	C.Surrounding = 5
	C.Fuzziness = 1

	if !concurent {
		test1b(engine, A, B, C) // one-by-one
	} else {
		test1c(engine, A, B, C) // concurent
	}
}

func main() {
	//test0()
	//test1(false)
	test1(true) // concurent
}
