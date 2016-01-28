package main

import (
	log "github.com/Sirupsen/logrus"

	"github.com/getryft/ryft-server/search"
	_ "github.com/getryft/ryft-server/search/ryfthttp"
	_ "github.com/getryft/ryft-server/search/ryftmux"
	_ "github.com/getryft/ryft-server/search/ryftprim"
)

func newRyftPrim() search.Engine {
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

	return engine
}

func newRyftHttp() search.Engine {
	backend := "ryfthttp"
	opts := map[string]interface{}{}
	engine, err := search.NewEngine(backend, opts)
	if err != nil {
		log.WithError(err).Fatalf("failed to get search engine")
	}
	log.WithField("name", backend).Infof("actual options: %+v", engine.Options())

	return engine
}

func main() {
	//test0()
	//test1(false) // step-by-step
	//test1(true) // concurent
	//test2() // HTTP
	//test3() // MUX

	//files1(false) // step-by-step
	//files2() // HTTP
	files3() // MUX
}
