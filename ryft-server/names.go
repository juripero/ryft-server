package main

import (
	"fmt"
	"path/filepath"
	"strconv"
)

var RyftoneMountPoint = "/ryftone" //TODO: from config
var ServerInstancePrefix = "RyftServer"

type Names struct {
	ResultFile, IdxFile string
}

var namesChan = make(chan Names, 256)

func StartNamesGenerator() {
	go func() {
		var s string
		for {
			for i := uint64(0); i <= ^uint64(0); i++ {
				s = strconv.FormatUint(i, 10)
				namesChan <- Names{"result-" + s + ".bin", "idx-" + s + ".txt"}
			}
		}
	}()
}

func GetNewNames() Names {
	return <-namesChan
}

func ResultsDirName() string {
	return fmt.Sprintf("%s-%d", ServerInstancePrefix, Port)
}

func ResultsDirPath(filenames ...string) string {
	return filepath.Join(append([]string{RyftoneMountPoint, ResultsDirName()}, filenames...)...)
}

func PathInRyftoneForResultDir(filenames ...string) string {
	return filepath.Join(append([]string{ResultsDirName()}, filenames...)...)
}
