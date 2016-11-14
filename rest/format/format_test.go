package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test format
func TestFormat(t *testing.T) {
	// create valid format
	check := func(name string, opts map[string]interface{}) {
		fmt, err := New(name, opts)
		assert.NoError(t, err)
		assert.NotNil(t, fmt)
	}

	// create bad format
	bad := func(name string, opts map[string]interface{}, expectedError string) {
		_, err := New(name, opts)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	check("json", nil)
	check("utf-8", nil)
	check("utf8", nil)
	check("null", nil)
	check("none", nil)
	check("raw", nil)
	check("xml", nil)

	bad("bad", nil, "is unsupported format")
}
