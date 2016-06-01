package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/getryft/ryft-server/search"
)

// run files locally
func runFiles1(log Logger, tag string, engine search.Engine, path string) {
	log("[%s]: start /files: %q", tag, path)

	res, err := engine.Files(path)
	if err != nil {
		log("[%s]: failed to start /files: %s", tag, err)
		panic(err)
	}

	log("[%s]: directory content: %v", tag, res)
}

// run multiple times: step-by-step
func runFilesStepByStep(log Logger, engine search.Engine, pathes ...string) {
	for i, path := range pathes {
		tag := fmt.Sprintf("S%d", i+1)
		runFiles1(log, tag, engine, path)
	}
}

// run multiple times: concurent
func runFilesConcurent(log Logger, engine search.Engine, pathes ...string) {
	wg := sync.WaitGroup{}

	// run each files in goroutine
	for i, path := range pathes {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()
			tag := fmt.Sprintf("X%d", i+1)
			runFiles1(log, tag, engine, path)
		}(i, path)

		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait() // wait all goroutines to finish
}
