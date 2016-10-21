package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// check dump a string
func testDumpAsString(t *testing.T, val interface{}, expected string) {
	s := DumpAsString(val)
	assert.Equal(t, expected, s, "bad dump string [%v]", val)
}

// dump as string cases
func TestDumpAsString(t *testing.T) {
	testDumpAsString(t, "", "")
	testDumpAsString(t, nil, "<nil>")
	testDumpAsString(t, []byte("hello"), "hello")
	testDumpAsString(t, []byte("\n\r\f"), "0x0a0d0c")
	testDumpAsString(t, []byte("привет"), "0xd0bfd180d0b8d0b2d0b5d182") // hello in russian, utf-8
}
