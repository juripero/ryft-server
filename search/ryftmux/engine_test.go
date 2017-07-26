package ryftmux

import (
	"fmt"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search/testfake"
	"github.com/stretchr/testify/assert"
)

var (
	testLogLevel = "error"
)

// set test log level
func testSetLogLevel() {
	SetLogLevelString(testLogLevel)
	testfake.SetLogLevelString(testLogLevel)
}

// create new fake engine
func newFake(records, errors int) *testfake.Engine {
	engine, _ := testfake.NewEngine(fmt.Sprintf("/tmp/ryft-%u", time.Now().UnixNano()), "ryftmux")
	engine.SearchReportRecords = records
	engine.SearchReportErrors = errors
	return engine
}

// test engine options
func TestEngineOptions(t *testing.T) {
	testSetLogLevel()

	assert.EqualValues(t, testLogLevel, GetLogLevel().String())

	backend := newFake(1, 0)
	engine, err := NewEngine(backend)
	assert.NoError(t, err)
	if assert.NotNil(t, engine) {
		assert.EqualValues(t, fmt.Sprintf("ryftmux{backends:[fake{home:%s/%s}]}", backend.MountPoint, backend.HomeDir), engine.String())
		assert.EqualValues(t, map[string]interface{}{
			"index-host": "",
		}, engine.Options())
	}
}
