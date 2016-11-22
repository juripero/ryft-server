package ryftdec

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/stretchr/testify/assert"
)

var (
	testLogLevel = "debug"
)

// fake engine to generate random data
type fakeEngine struct {
	Host       string
	MountPoint string
	HomeDir    string
	Instance   string

	// report to /search
	ErrorForSearch error
	ReportLatency  time.Duration

	// list of search done
	searchDone []*search.Config

	// report to /files
	FilesToReport []string
	DirsToReport  []string
	ErrorForFiles error
}

// create new fake engine
func newFake(records, errors int) *fakeEngine {
	return &fakeEngine{
		MountPoint: "/tmp",
		HomeDir:    "/ryft-test/",
		Instance:   ".work",
	}
}

// get string
func (fe fakeEngine) String() string {
	return fmt.Sprintf("fake{}")
}

// Get current engine options.
func (fe *fakeEngine) Options() map[string]interface{} {
	return map[string]interface{}{
		"instance-name": fe.Instance,
		"home-dir":      fe.HomeDir,
		"ryftone-mount": fe.MountPoint,
	}
}

// Run *synchronous* "/files" operation.
func (fe *fakeEngine) Files(path string) (*search.DirInfo, error) {
	info := search.NewDirInfo(path)
	info.AddFile(fe.FilesToReport...)
	info.AddDir(fe.DirsToReport...)
	return info, fe.ErrorForFiles
}

// cleanup all working directories
func (fe *fakeEngine) cleanup() {
	os.RemoveAll(filepath.Join(fe.MountPoint, fe.HomeDir))
}

// Check multiplexing of files and directories
func TestEngineFiles(t *testing.T) {
	SetLogLevelString(testLogLevel)

	f1 := newFake(0, 0)
	f1.FilesToReport = []string{"1.txt", "2.txt"}
	f1.DirsToReport = []string{"a", "b"}

	// valid (usual case)
	engine, err := NewEngine(f1, -1, false)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		info, err := engine.Files("foo")
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

	engine, err := NewEngine(newFake(1, 0), -1, false)
	assert.NoError(t, err)
	if assert.NotNil(t, engine) {
		assert.EqualValues(t, "ryftdec{backend:fake{}}", engine.String())
		assert.EqualValues(t, map[string]interface{}{
			"instance-name": ".work",
			"home-dir":      "/ryft-test/",
			"ryftone-mount": "/tmp",
		}, engine.Options())
	}
}
