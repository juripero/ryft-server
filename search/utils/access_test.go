package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test ParseField
func TestParseField(t *testing.T) {
	// check
	check_ex := func(field string, str2idx []string, idx2str []string, expected ...interface{}) {
		f, err := ParseField(field)
		if assert.NoError(t, err) {
			f = f.StringToIndex(str2idx)
			f = f.IndexToString(idx2str)

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
	check := func(field string, expected ...interface{}) {
		check_ex(field, nil, nil, expected...)
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
	check("[5]", 4)
	check("a.[5]", "a", 4)
	check("[5].b", 4, "b")
	check_ex("a.b.c", []string{"x", "b", "z"}, nil, "a", 1, "c")
	check_ex("a.[2].c", nil, []string{"x", "b", "z"}, "a", "b", "c")

	bad("[11111111111111111111111]", "failed to parse field index")
	bad("[5[", "found instead of ]")
	bad("[xyz]", "found instead of index")
	bad("(-)", "unexpected token found")

	tmp := Field{fieldStr("a"), fieldInt(4), fieldStr("b")}
	assert.EqualValues(t, tmp.String(), "a.[5].b")
	assert.EqualValues(t, MakeIntField(554).String(), "[555]")
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
	check([]string{"a", "b", "c"}, "[2]", "b")
	check([]interface{}{5, "b", false}, "[1]", 5)
	check([]interface{}{5, "b", false}, "[2]", "b")
	check([]interface{}{5, "b", true}, "[3]", true)
	bad([]string{"a", "b", "c"}, "[100]", "requested value is missed")
	bad([]interface{}{5, "b", false}, "[100]", "requested value is missed")
	bad("foo/bar", "[1]", "bad data type for index field")
}
