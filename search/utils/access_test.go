package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test nested access
func TestAccessValue(t *testing.T) {
	// check
	check := func(data interface{}, field string, expected interface{}) {
		v, err := AccessValue(data, field)
		if assert.NoError(t, err) {
			assert.EqualValues(t, expected, v)
		}
	}

	check("hello", ".", "hello")
	check(nil, "", nil)
	check(map[string]interface{}{
		"foo": "bar",
	}, "foo", "bar")
	check(map[string]interface{}{
		"foo": "bar",
	}, ".foo", "bar")
	check(map[string]interface{}{
		"foo": "bar",
	}, "foo.", "bar")
	check(map[string]interface{}{
		"foo": map[interface{}]interface{}{
			"bar": "hello",
		},
	}, "foo.bar", "hello")
}
