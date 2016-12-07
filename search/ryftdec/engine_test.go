package ryftdec

import (
	"sort"
	"testing"

	"github.com/getryft/ryft-server/search/testfake"
	"github.com/stretchr/testify/assert"
)

var (
	testLogLevel = "error"
)

// create new fake engine
func newFake(records, errors int) *testfake.Engine {
	engine, _ := testfake.NewEngine("/tmp", "/ryft")
	return engine
}

// Check multiplexing of files and directories
func TestEngineFiles(t *testing.T) {
	SetLogLevelString(testLogLevel)

	f1 := newFake(0, 0)
	f1.FilesReportFiles = []string{"1.txt", "2.txt"}
	f1.FilesReportDirs = []string{"a", "b"}

	// valid (usual case)
	engine, err := NewEngine(f1, -1, false, false)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		info, err := engine.Files("foo", false)
		if assert.NoError(t, err) && assert.NotNil(t, info) {
			assert.EqualValues(t, "foo", info.Path)

			sort.Strings(info.Files)
			assert.EqualValues(t, []string{"1.txt", "2.txt"}, info.Files)

			sort.Strings(info.Dirs)
			assert.EqualValues(t, []string{"a", "b"}, info.Dirs)
		}
	}
}

// test engine options
func TestEngineOptions(t *testing.T) {
	SetLogLevelString(testLogLevel)

	assert.EqualValues(t, testLogLevel, GetLogLevel().String())

	engine, err := NewEngine(newFake(1, 0), -1, false, false)
	assert.NoError(t, err)
	if assert.NotNil(t, engine) {
		assert.EqualValues(t, "ryftdec{backend:fake{home:/tmp/ryft}}", engine.String())
		assert.EqualValues(t, map[string]interface{}{
			"instance-name": ".work",
			"home-dir":      "/ryft",
			"ryftone-mount": "/tmp",
			"host-name":     "",
		}, engine.Options())
	}
}
