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
	engine, err := NewEngine(f1, nil)
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
	if err := SetLogLevelString("BAD"); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "not a valid logrus Level")
	}

	// check for good case
	check := func(opts map[string]interface{}) {
		if engine, err := NewEngine(newFake(0, 0), opts); assert.NoError(t, err) {
			assert.EqualValues(t, opts, engine.Options())
		}
	}
	check2 := func(opts map[string]interface{}, expectedOpts map[string]interface{}) {
		if engine, err := NewEngine(newFake(0, 0), opts); assert.NoError(t, err) {
			assert.EqualValues(t, expectedOpts, engine.Options())
		}
	}

	// check for bad case
	bad := func(opts map[string]interface{}, expectedError string) {
		if _, err := NewEngine(newFake(0, 0), opts); assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	// get fake options
	fake := func(name string, val interface{}) map[string]interface{} {
		opts := map[string]interface{}{
			"instance-name":            ".work",
			"ryftone-mount":            "/tmp",
			"home-dir":                 "/ryft",
			"host-name":                "",
			"compat-mode":              false,
			"optimizer-limit":          -1,
			"optimizer-do-not-combine": "",
		}

		if len(name) != 0 {
			opts[name] = val
		}

		return opts
	}

	// check default options
	engine, err := NewEngine(newFake(1, 0), nil)
	assert.NoError(t, err)
	if assert.NotNil(t, engine) {
		assert.EqualValues(t, "ryftdec{backend:fake{home:/tmp/ryft}}", engine.String())
		assert.EqualValues(t, map[string]interface{}{
			"instance-name":            ".work",
			"ryftone-mount":            "/tmp",
			"home-dir":                 "/ryft",
			"host-name":                "",
			"compat-mode":              false,
			"optimizer-limit":          -1,
			"optimizer-do-not-combine": "",
		}, engine.Options())
	}

	check(fake("compat-mode", true))
	// check(fake("keep-files", true))
	check(fake("optimizer-limit", 10))
	check(fake("optimizer-limit", -1))
	check(fake("optimizer-do-not-combine", "es"))
	check(fake("optimizer-do-not-combine", "fhs"))
	check(fake("optimizer-do-not-combine", "ds:ts"))

	check2(fake("optimizer-do-not-combine", "ds ts"), fake("optimizer-do-not-combine", "ds:ts"))
	check2(fake("optimizer-do-not-combine", "ds,ts"), fake("optimizer-do-not-combine", "ds:ts"))
	check2(fake("optimizer-do-not-combine", "ds;ts"), fake("optimizer-do-not-combine", "ds:ts"))
	check2(fake("optimizer-do-not-combine", "ds  ts"), fake("optimizer-do-not-combine", "ds:ts"))
	check2(fake("optimizer-do-not-combine", "  ds ,,;;::,, ts  "), fake("optimizer-do-not-combine", "ds:ts"))

	bad(fake("compat-mode", []byte{}), `failed to parse "compat-mode"`)
	bad(fake("keep-files", []byte{}), `failed to parse "keep-files"`)
	bad(fake("optimizer-limit", "bad"), `failed to parse "optimizer-limit"`)
	bad(fake("optimizer-do-not-combine", false), `failed to parse "optimizer-do-not-combine"`)
}
