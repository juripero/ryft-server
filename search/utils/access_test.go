package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test ParseField
func TestParseField(t *testing.T) {
	// check
	check := func(field string, expected ...interface{}) {
		f, err := ParseField(field)
		if assert.NoError(t, err) {
			var ef Field
			for _, e := range expected {
				switch v := e.(type) {
				case int:
					ef = append(ef, fieldInt(v))
				case string:
					ef = append(ef, fieldStr(v))
				default:
					assert.Fail(t, "unexpected type")
				}
			}
			assert.EqualValues(t, ef, f)
		}
	}

	// parse a "bad" map
	bad := func(field string, expectedError string) {
		_, err := ParseField(field)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", field)
		}
	}

	check("")       // no field
	check("a", "a") // string field
	check("a.b", "a", "b")
	check("..a...b..", "a", "b")
	check(`"a b"."c"`, "a b", "c")
	check("1", "1")
	check("[5]", 5)
	check("a.[5]", "a", 5)
	check("[5].b", 5, "b")

	bad("[11111111111111111111111]", "failed to parse field index")
	bad("[5[", "found instead of ]")
	bad("[xyz]", "found instead of index")
	bad("(-)", "unexpected token found")

	tmp := Field{fieldStr("a"), fieldInt(5), fieldStr("b")}
	assert.EqualValues(t, tmp.String(), "a.[5].b")
}

// test nested field access
func TestAccessValue(t *testing.T) {
	// check
	check := func(data interface{}, field string, expected interface{}) {
		f, err := ParseField(field)
		if !assert.NoError(t, err) {
			return
		}

		if v, err := f.GetValue(data); assert.NoError(t, err) {
			assert.EqualValues(t, expected, v)
		}
	}

	// bad
	bad := func(data interface{}, field string, expectedError string) {
		f, err := ParseField(field)
		if !assert.NoError(t, err) {
			return
		}

		if _, err := f.GetValue(data); assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	check(nil, "", nil)
	check("hello", "", "hello")
	check("hello", ".", "hello")
	check("hello", "..", "hello")

	// map
	check(map[string]interface{}{
		"foo": "bar",
	}, "foo", "bar")
	check(map[interface{}]interface{}{
		"foo": "bar",
	}, ".foo", "bar")
	check(map[string]interface{}{
		"foo": map[interface{}]interface{}{
			"bar": "hello",
		},
	}, "foo.bar", "hello")
	bad(map[string]interface{}{
		"foo": "bar",
	}, "bar", "requested value is missed")
	bad(map[interface{}]interface{}{
		"foo": "bar",
	}, ".bar", "requested value is missed")
	bad("foo/bar", "foo.", "bad data type for string field")

	// array
	check([]string{"a", "b", "c"}, "[1]", "b")
	check([]interface{}{5, "b", false}, "[0]", 5)
	check([]interface{}{5, "b", false}, "[1]", "b")
	check([]interface{}{5, "b", true}, "[2]", true)
	bad([]string{"a", "b", "c"}, "[100]", "requested value is missed")
	bad([]interface{}{5, "b", false}, "[100]", "requested value is missed")
	bad("foo/bar", "[0]", "bad data type for index field")
}
