package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// dump as string cases
func TestDumpAsString(t *testing.T) {
	// check dump a string
	check := func(val interface{}, expected string) {
		s := DumpAsString(val)
		assert.Equal(t, expected, s, "bad dump string [%v]", val)
	}

	check("", "")
	check(nil, "<nil>")
	check([]byte("hello"), "hello")
	check([]byte("\n\r\f"), "#0a0d0c")
	check([]byte("привет"), "#d0bfd180d0b8d0b2d0b5d182") // hello in russian, utf-8
}
