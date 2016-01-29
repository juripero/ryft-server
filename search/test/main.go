package main

import (
	stdlog "log"
	"os"
	"runtime/pprof"

	// import for side effects
	"github.com/getryft/ryft-server/search"
	_ "github.com/getryft/ryft-server/search/ryfthttp"
	_ "github.com/getryft/ryft-server/search/ryftmux"
	_ "github.com/getryft/ryft-server/search/ryftprim"
)

var (
	log Logger = stdlog.Printf
)

// logger function is used as an abstraction to unify
// access to `log.Printf` and `t.Logf` from unit tests
type Logger func(format string, args ...interface{})

// application entry point
func main() {
	defer func() {
		if e := recover(); e != nil {
			stdlog.Fatalf("FAILED")
		}
	}()

	// enable profiling
	if true {
		f, err := os.Create("search.prof")
		if err != nil {
			log("failed to create profile file: %s", err)
			panic(err)
		}

		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	ryftprimLogLevel = "debug"
	ryfthttpLogLevel = "debug"

	// printSearchEngines()

	search1(false) // step-by-step
	//search1(true) // concurent
	// test2() // HTTP
	//test3() // MUX

	//files1(false) // step-by-step
	//files2() // HTTP
	//files3() // MUX
}

// ryftprim seach
func search1(concurent bool) {
	engine := newRyftPrim(log)

	// plain texts
	A := search.NewConfig(`10`, "/regression/passengers.txt")
	B := search.NewConfig(`20`, "/regression/passengers.txt")
	C := search.NewConfig(`555`, "/regression/passengers.txt")
	C.Fuzziness = 1

	// XML records
	D := search.NewConfig(`(RECORD.id CONTAINS "1003")`, "/regression/*.pcrime")
	E := search.NewConfig(`(RECORD.id CONTAINS "1003100")`, "/regression/*.pcrime")
	F := search.NewConfig(`(RECORD.desc CONTAINS "VEHICLE")`, "/regression/*.pcrime")

	_, _, _, _, _, _ = A, B, C, D, E, F

	cfgs := []search.Config{*D}
	//cfgs = append(cfgs, *A, *B, *C)
	//cfgs = append(cfgs, *D, *E, *F)
	//cfgs = append(cfgs, *A, *B, *C)
	//cfgs = append(cfgs, *A, *B, *C)
	//cfgs = append(cfgs, *A, *B, *C)
	//cfgs = append(cfgs, *A, *B, *C)

	if !concurent {
		runSearchOneByOne(log, engine, cfgs...)
	} else {
		runSearchConcurent(log, engine, cfgs...)
	}
}

// ryfthttp search
func search2() {
	engine := newRyftHttp(log)

	//A := search.NewConfig(`"test"`, "/regression/*.txt")
	A := search.NewConfig(`(RECORD.desc CONTAINS "VEHICLE")`, "*.pcrime")
	A.Surrounding = 10
	A.Fuzziness = 0

	runSearch1(log, "H1", engine, A)
}

// ryftmux search
func search3() {
	// tripple seach
	engine := newRyftMux(log,
		newRyftPrim(log),
		newRyftPrim(log),
		newRyftPrim(log),
	)

	A := search.NewConfig(`"test"`, "/regression/*.txt")
	A.Surrounding = 10
	A.Fuzziness = 1

	runSearch1(log, "M1", engine, A)
}

// ryftprim files
func files1(concurent bool) {
	engine := newRyftPrim(log)

	paths := []string{}
	paths = append(paths, "/")
	paths = append(paths, "/regression")

	if !concurent {
		runFilesStepByStep(log, engine, paths...) // one-by-one
	} else {
		runFilesConcurent(log, engine, paths...) // concurent
	}
}

// ryfthttp files
func files2() {
	engine := newRyftHttp(log)

	runFiles1(log, "H1", engine, "/")
}

// ryftmux files
func files3() {
	// tripple seach
	engine := newRyftMux(log,
		newRyftPrim(log),
		newRyftPrim(log),
		newRyftPrim(log),
	)

	runFiles1(log, "M1", engine, "/")
}
