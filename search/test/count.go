package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/getryft/ryft-server/search"
)

// run one count
func runCount1(log Logger, tag string, engine search.Engine, cfg *search.Config) SearchResult {
	log("[%s]: start /count: %s", tag, cfg)

	res, err := engine.Count(cfg)
	if err != nil {
		log("[%s]: failed to start /count: %s", tag, err)
		panic(err)
	}

	return grabResults(log, tag, res, false)
}

// run multiple count: step-by-step
func runCountOneByOne(log Logger, engine search.Engine, cfgs ...search.Config) {
	for i, cfg := range cfgs {
		tag := fmt.Sprintf("S%d", i+1)
		runCount1(log, tag, engine, &cfg)
	}
}

// run multiple count: concurent
func runCountConcurent(log Logger, engine search.Engine, cfgs ...search.Config) {
	wg := sync.WaitGroup{}

	// run each count in goroutine
	for i, cfg := range cfgs {
		wg.Add(1)
		go func(i int, cfg search.Config) {
			defer wg.Done()
			tag := fmt.Sprintf("X%d", i+1)
			runCount1(log, tag, engine, &cfg)
		}(i, cfg)

		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait() // wait all goroutines to finish
}
