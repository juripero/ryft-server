package ryftmux

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	testLogLevel = "error"
)

// fake engine to generate random data
type fakeEngine struct {
	Host string

	// report to /search
	RecordsToReport int
	ErrorsToReport  int
	ErrorForSearch  error
	ReportLatency   time.Duration

	// report to /files
	FilesToReport []string
	DirsToReport  []string
	ErrorForFiles error
	PathSuffix    string
}

// create new fake engine
func newFake(records, errors int) *fakeEngine {
	return &fakeEngine{
		RecordsToReport: records,
		ErrorsToReport:  errors,
	}
}

// get string
func (fe fakeEngine) String() string {
	return fmt.Sprintf("fake{%d,%d}",
		fe.RecordsToReport,
		fe.ErrorsToReport)
}

// Get current engine options.
func (fe *fakeEngine) Options() map[string]interface{} {
	return map[string]interface{}{
		"records": fe.RecordsToReport,
		"errors":  fe.ErrorsToReport,
	}
}

// test engine options
func TestEngineOptions(t *testing.T) {
	SetLogLevelString(testLogLevel)

	assert.EqualValues(t, testLogLevel, GetLogLevel().String())

	engine, err := NewEngine(newFake(1, 0))
	assert.NoError(t, err)
	if assert.NotNil(t, engine) {
		assert.EqualValues(t, "ryftmux{backends:[fake{1,0}]}", engine.String())
		assert.EqualValues(t, map[string]interface{}{
			"index-host": "",
		}, engine.Options())
	}
}
