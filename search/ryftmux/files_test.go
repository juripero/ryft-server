package ryftmux

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Check multiplexing of files and directories
func TestEngineFiles(t *testing.T) {
	SetLogLevelString(testLogLevel)

	f1 := newFake(0, 0)
	f1.FilesReportFiles = []string{"1.txt", "2.txt"}
	f1.FilesReportDirs = []string{"a", "b"}

	f2 := newFake(0, 0)
	f2.FilesReportFiles = []string{"2.txt", "3.txt"}
	f2.FilesReportDirs = []string{"b", "c"}

	f3 := newFake(0, 0)
	f3.FilesReportFiles = []string{"3.txt", "4.txt"}
	f3.FilesReportDirs = []string{"c", "d"}

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
	f1.FilesReportError = fmt.Errorf("disabled")
	if assert.Error(t, f1.FilesReportError) {
		info, err := engine.Files("foo")
		if assert.NoError(t, err) && assert.NotNil(t, info) {
			assert.EqualValues(t, "foo", info.Path)

			sort.Strings(info.Files)
			assert.EqualValues(t, []string{ /*"1.txt",*/ "2.txt", "3.txt", "4.txt"}, info.Files)

			sort.Strings(info.Dirs)
			assert.EqualValues(t, []string{ /*"a", */ "b", "c", "d"}, info.Dirs)
		}
	}
	f1.FilesReportError = nil

	// path inconsistency
	f1.FilesPathSuffix = "-1"
	if assert.NoError(t, f1.FilesReportError) {
		info, err := engine.Files("foo")
		if assert.Error(t, err) && assert.Nil(t, info) {
			assert.Contains(t, err.Error(), "inconsistent path")
		}
	}
	f1.FilesPathSuffix = ""
}
