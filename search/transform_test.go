package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test regexp-match transformation
func TestRegexpMatch(t *testing.T) {
	// good case
	check := func(expr string, in string, expectedOut string, expectedSkip bool) {
		tx, err := NewRegexpMatch(expr)
		if !assert.NoError(t, err) {
			return
		}

		out, skip, err := tx.Process([]byte(in))
		if assert.NotNil(t, out) && assert.NoError(t, err) {
			assert.EqualValues(t, expectedOut, out)
			assert.EqualValues(t, expectedSkip, skip)
		}
	}

	// bad case
	bad := func(expr string, in string, expectedError string) {
		tx, err := NewRegexpMatch(expr)
		if err != nil {
			assert.Contains(t, err.Error(), expectedError)
			return
		}

		out, _, err := tx.Process([]byte(in))
		if assert.Nil(t, out) && assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	check("$apple^", "hello", "hello", true)
	check("^.*ell.*$", "hello", "hello", false)
	bad("$(", "hello", "error parsing regexp")
}

// test regexp-replace transformation
func TestRegexpReplace(t *testing.T) {
	// good case
	check := func(expr, templ string, in string, expectedOut string, expectedSkip bool) {
		tx, err := NewRegexpReplace(expr, templ)
		if !assert.NoError(t, err) {
			return
		}

		out, skip, err := tx.Process([]byte(in))
		if assert.NotNil(t, out) && assert.NoError(t, err) {
			assert.EqualValues(t, expectedOut, out)
			assert.EqualValues(t, expectedSkip, skip)
		}
	}

	// bad case
	bad := func(expr, templ string, in string, expectedError string) {
		tx, err := NewRegexpReplace(expr, templ)
		if err != nil {
			assert.Contains(t, err.Error(), expectedError)
			return
		}

		out, _, err := tx.Process([]byte(in))
		if assert.Nil(t, out) && assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	check("$apple^", "$1", "hello", "hello", false) // keep as is
	check("^.*(ell).*$", "Z$1", "hello", "Zell", false)
	bad("$(", "$1", "hello", "error parsing regexp")
}

// test script-call transformation
func TestScriptCall(t *testing.T) {
	// good case
	check := func(script []string, in string, expectedOut string, expectedSkip bool) {
		tx, err := NewScriptCall(script, "")
		if !assert.NoError(t, err) {
			return
		}

		out, skip, err := tx.Process([]byte(in))
		if assert.NotNil(t, out) && assert.NoError(t, err) {
			assert.EqualValues(t, expectedOut, out)
			assert.EqualValues(t, expectedSkip, skip)
		}
	}

	// bad case
	bad := func(script []string, in string, expectedError string) {
		tx, err := NewScriptCall(script, "")
		if err != nil {
			assert.Contains(t, err.Error(), expectedError)
			return
		}

		out, _, err := tx.Process([]byte(in))
		if assert.Nil(t, out) && assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	check([]string{"/bin/cat"}, "hello", "hello", false)
	check([]string{"/bin/grep", "hell"}, "apple\nhello\norange\n", "hello\n", false)
	check([]string{"/bin/false"}, "hello", "hello", true)
	bad([]string{}, "hello", "no script path provided")
	bad([]string{"/bin/missing-script"}, "hello", "no valid script found")
}
