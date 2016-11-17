package ryftmux

import (
	"fmt"
	"sort"
	"testing"

	"github.com/getryft/ryft-server/search"
	"github.com/stretchr/testify/assert"
)

// Run *synchronous* "/files" operation.
func (fe *fakeEngine) Files(path string) (*search.DirInfo, error) {
	info := search.NewDirInfo(path + fe.PathSuffix)
	info.AddFile(fe.FilesToReport...)
	info.AddDir(fe.DirsToReport...)
	return info, fe.ErrorForFiles
}

// Check multiplexing of files and directories
func TestEngineFiles(t *testing.T) {
	SetLogLevelString(testLogLevel)

	f1 := newFake(0, 0)
	f1.FilesToReport = []string{"1.txt", "2.txt"}
	f1.DirsToReport = []string{"a", "b"}

	f2 := newFake(0, 0)
	f2.FilesToReport = []string{"2.txt", "3.txt"}
	f2.DirsToReport = []string{"b", "c"}

	f3 := newFake(0, 0)
	f3.FilesToReport = []string{"3.txt", "4.txt"}
	f3.DirsToReport = []string{"c", "d"}

	// valid (usual case)
	engine, err := NewEngine(f1, f2, f3)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		info, err := engine.Files("foo")
		if assert.NoError(t, err) && assert.NotNil(t, info) {
			assert.EqualValues(t, "foo", info.Path)

			sort.Strings(info.Files)
			assert.EqualValues(t, []string{"1.txt", "2.txt", "3.txt", "4.txt"}, info.Files)

			sort.Strings(info.Dirs)
			assert.EqualValues(t, []string{"a", "b", "c", "d"}, info.Dirs)
		}
	}

	// one backend fail
	f1.ErrorForFiles = fmt.Errorf("disabled")
	if assert.Error(t, f1.ErrorForFiles) {
		info, err := engine.Files("foo")
		if assert.NoError(t, err) && assert.NotNil(t, info) {
			assert.EqualValues(t, "foo", info.Path)

			sort.Strings(info.Files)
			assert.EqualValues(t, []string{ /*"1.txt",*/ "2.txt", "3.txt", "4.txt"}, info.Files)

			sort.Strings(info.Dirs)
			assert.EqualValues(t, []string{ /*"a", */ "b", "c", "d"}, info.Dirs)
		}
	}
	f1.ErrorForFiles = nil

	// path inconsistency
	f1.PathSuffix = "-1"
	if assert.NoError(t, f1.ErrorForFiles) {
		info, err := engine.Files("foo")
		if assert.Error(t, err) && assert.Nil(t, info) {
			assert.Contains(t, err.Error(), "inconsistent path")
		}
	}
	f1.PathSuffix = ""
}
