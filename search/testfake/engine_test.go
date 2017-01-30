package testfake

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testLogLevel = "error"
)

// test engine options
func TestEngineOptions(t *testing.T) {
	SetLogLevelString(testLogLevel)

	assert.EqualValues(t, testLogLevel, GetLogLevel().String())

	engine, err := NewEngine("/tmp", "/ryft")
	assert.NoError(t, err)
	if assert.NotNil(t, engine) {
		assert.EqualValues(t, "fake{home:/tmp/ryft}", engine.String())
		assert.EqualValues(t, map[string]interface{}{
			"instance-name": ".work",
			"home-dir":      "/ryft",
			"ryftone-mount": "/tmp",
			"host-name":     "",
		}, engine.Options())
	}
}
