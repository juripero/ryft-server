package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// create valid format
func testFormatNew(t *testing.T, name string, opts map[string]interface{}) {
	fmt, err := New(name, opts)
	assert.NoError(t, err)
	assert.NotNil(t, fmt)
}

// create bad format
func testFormatNewBad(t *testing.T, name string, opts map[string]interface{}, expectedError string) {
	_, err := New(name, opts)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), expectedError)
	}
}

// test format
func TestFormat(t *testing.T) {
	testFormatNew(t, "json", nil)
	testFormatNew(t, "utf-8", nil)
	testFormatNew(t, "utf8", nil)
	testFormatNew(t, "null", nil)
	testFormatNew(t, "none", nil)
	testFormatNew(t, "raw", nil)
	testFormatNew(t, "xml", nil)

	testFormatNewBad(t, "bad", nil, "is unsupported format")
}
