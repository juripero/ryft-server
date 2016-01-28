package main

import (
	"fmt"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftmux"
)

// run files locally
func files1a(tag string, engine search.Engine, path string) {
	xlog := log.WithField("tag", tag)
	xlog.Printf("start /files: %q", path)

	res, err := engine.Files(path)
	if err != nil {
		xlog.WithError(err).Fatalf("failed to start /files")
	}

	xlog.Infof("directory info: %v", res)
}

// run multiple times: step-by-step
func files1b(engine search.Engine, paths ...string) {
	for i, path := range paths {
		files1a(fmt.Sprintf("FS%d", i+1), engine, path)
	}
}

// run multiple times: concurent
func files1c(engine search.Engine, paths ...string) {
	wg := sync.WaitGroup{}

	// run each search in goroutine
	for i, path := range paths {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()
			files1a(fmt.Sprintf("FX%d", i+1), engine, path)
		}(i, path)

		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait() // wait all goroutines to finish
}

func files1(concurent bool) {
	engine := newRyftPrim()

	paths := []string{}
	paths = append(paths, "/")
	paths = append(paths, "/regression")

	if !concurent {
		files1b(engine, paths...) // one-by-one
	} else {
		files1c(engine, paths...) // concurent
	}
}

func files2() {
	engine := newRyftHttp()

	files1a("FH1", engine, "/")
}

func files3() {
	// tripple seach
	engine, err := ryftmux.NewEngine(
		newRyftPrim(),
		newRyftPrim(),
		newRyftPrim(),
	)
	if err != nil {
		log.WithError(err).Fatalf("failed to get search engine")
	}

	files1a("FM1", engine, "/")
}
