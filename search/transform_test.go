package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test call script
func TestCallScript(t *testing.T) {
	// good case
	check := func(script []string, in string, expectedOut string, expectedSkip bool) {
		tx, err := NewScriptCall(script, "")
		if !assert.NoError(t, err) {
			return
		}

		out, skip, err := tx.Transform([]byte(in))
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

		out, _, err := tx.Transform([]byte(in))
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
