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

// hex escape
func TestHexEscape(t *testing.T) {
	// check hex escape
	check := func(val []byte, expected string) {
		s := HexEscape(val)
		assert.Equal(t, expected, s)
	}

	check(nil, "")
	check([]byte{}, "")
	check([]byte{0x0a, 0x0d, 0x0c}, `\x0a\x0d\x0c`)
	check([]byte("\n\r\f"), `\x0a\x0d\x0c`)
}
