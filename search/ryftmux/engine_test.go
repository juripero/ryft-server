package ryftmux

import (
	"testing"

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
	engine, _ := testfake.NewEngine("/tmp", "ryft-mux")
	engine.SearchReportRecords = records
	engine.SearchReportErrors = errors
	return engine
}

// test engine options
func TestEngineOptions(t *testing.T) {
	testSetLogLevel()

	assert.EqualValues(t, testLogLevel, GetLogLevel().String())

	engine, err := NewEngine(newFake(1, 0))
	assert.NoError(t, err)
	if assert.NotNil(t, engine) {
		assert.EqualValues(t, "ryftmux{backends:[fake{home:/tmp/ryft-mux}]}", engine.String())
		assert.EqualValues(t, map[string]interface{}{
			"index-host": "",
		}, engine.Options())
	}
}
