package csv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test format options
func TestFormatOptions(t *testing.T) {
	// bad option type
	_, err := New(map[string]interface{}{
		"fields": 555,
	})
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "unsupported option type")
	}

	// fields from string
	fmt1, err := New(map[string]interface{}{
		"fields": "a,b",
	})
	if assert.NoError(t, err) && assert.NotNil(t, fmt1) {
		assert.EqualValues(t, fmt1.Fields, []string{"a", "b"})
	}

	// fields from []string
	fmt2, err := New(map[string]interface{}{
		"fields": []string{"a", "b"},
	})
	if assert.NoError(t, err) && assert.NotNil(t, fmt2) {
		assert.EqualValues(t, fmt2.Fields, []string{"a", "b"})
	}

	// AddFields
	fmt2.AddFields("c,d")
	assert.EqualValues(t, fmt2.Fields, []string{"a", "b", "c", "d"})
}
