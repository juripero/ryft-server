package ryftdec

import (
	// "fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test extension detection
func TestDetectExtension(t *testing.T) {
	// good case
	check := func(fileNames []string, dataOut string, expected string) {
		ext, err := detectExtension(fileNames, dataOut)
		if assert.NoError(t, err) {
			assert.Equal(t, expected, ext)
		}
	}

	// bad case
	bad := func(fileNames []string, dataOut string, expectedError string) {
		_, err := detectExtension(fileNames, dataOut)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	check([]string{}, "out.txt", ".txt")
	check([]string{"a.txt"}, "", ".txt")
	check([]string{"a.txt", "b.txt"}, "", ".txt")
	check([]string{"a.dat", "b.dat"}, "", ".dat")
	bad([]string{"a.txt", "b.dat"}, "", "ambiguous extension")
	bad([]string{"a.txt", "b.dat"}, "c.jpeg", "ambiguous extension")
	check([]string{}, "", "")
	check([]string{"foo/a.txt", "my.test/b.txt"}, "", ".txt")
	check([]string{"foo/a.txt", "my.test/b.txt"}, "data.txt", ".txt")
	check([]string{"foo/*.txt", "my.test/*txt"}, "", ".txt")
	check([]string{"foo/*.txt", "my.test/*"}, "data.txt", ".txt")
	check([]string{"my.test/*"}, "data.txt", ".txt")
	check([]string{"nyctaxi/xml/2015/yellow/*"}, "ryftnyctest.nxml", ".nxml")
}

// test
func TestRelativeToHome(t *testing.T) {
	assert.EqualValues(t, "dir", relativeToHome("/ryftone", "/ryftone/dir"))
	assert.EqualValues(t, "dir", relativeToHome("/ryftone", "dir")) // fallback
}
