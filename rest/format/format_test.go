package format

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test format
func TestFormat(t *testing.T) {
	// create valid format
	check := func(name string, opts map[string]interface{}, isNull bool) {
		fmt, err := New(name, opts)
		assert.NoError(t, err)
		assert.NotNil(t, fmt)
		assert.EqualValues(t, isNull, IsNull(name))
	}

	// create bad format
	bad := func(name string, opts map[string]interface{}, expectedError string) {
		_, err := New(name, opts)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	check("json", nil, false)
	check("utf-8", nil, false)
	check("utf8", nil, false)
	check("null", nil, true)
	check("none", nil, true)
	check("raw", nil, false)
	check("xml", nil, false)
	check("csv", nil, false)

	bad("bad", nil, "is unsupported format")
}
