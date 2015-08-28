// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package names

import (
	"fmt"
	"path/filepath"
	"strconv"
)

var RyftoneMountPoint = "/ryftone"
var ServerInstancePrefix = "RyftServer"
var Port = 8765

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

func New() Names {
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
