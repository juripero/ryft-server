package testfake

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	testLogLevel = "error"
)

// test engine options
func TestEngineOptions(t *testing.T) {
	SetLogLevelString(testLogLevel)

	assert.EqualValues(t, testLogLevel, GetLogLevel().String())

	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	engine, err := NewEngine(root, "/test")
	assert.NoError(t, err)
	if assert.NotNil(t, engine) {
		assert.EqualValues(t, fmt.Sprintf("fake{home:%s/test}", root), engine.String())
		assert.EqualValues(t, map[string]interface{}{
			"instance-name": ".work",
			"home-dir":      "/test",
			"ryftone-mount": root,
			"host-name":     "",
		}, engine.Options())
	}
}
