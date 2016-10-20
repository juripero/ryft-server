package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// parse a good index line
func testParseIndexGood(t *testing.T, line string, filename string, offset, length uint64, dist uint8) {
	idx, err := ParseIndex([]byte(line))
	if assert.NoError(t, err) {
		assert.Equal(t, filename, idx.File, "bad filename in [%s]", line)
		assert.Equal(t, offset, idx.Offset, "bad offset in [%s]", line)
		assert.Equal(t, length, idx.Length, "bad length in [%s]", line)
		assert.Equal(t, dist, idx.Fuzziness, "bad fuzziness in [%s]", line)
		assert.Equal(t, "", idx.Host, "no host expected", line)
	}
}

// parse a "bad" index line
func testParseIndexBad(t *testing.T, line string, expectedError string) {
	_, err := ParseIndex([]byte(line))
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), expectedError, "unexpected error [%s]", line)
	}
}

// good cases
func TestParseIndexGood(t *testing.T) {
	testParseIndexGood(t, "foo.txt,1,2,3", "foo.txt", 1, 2, 3)

	// spaces in filename
	testParseIndexGood(t, " foo.txt,0,0,0", "foo.txt", 0, 0, 0)
	testParseIndexGood(t, "foo.txt ,0,0,0", "foo.txt", 0, 0, 0)
	testParseIndexGood(t, " foo.txt ,0,0,0", "foo.txt", 0, 0, 0)

	// spaces in offset
	testParseIndexGood(t, "foo.txt, 1,2,3", "foo.txt", 1, 2, 3)
	testParseIndexGood(t, "foo.txt,1 ,2,3", "foo.txt", 1, 2, 3)
	testParseIndexGood(t, "foo.txt, 1 ,2,3", "foo.txt", 1, 2, 3)

	// spaces in length
	testParseIndexGood(t, "foo.txt,1, 2,3", "foo.txt", 1, 2, 3)
	testParseIndexGood(t, "foo.txt,1,2 ,3", "foo.txt", 1, 2, 3)
	testParseIndexGood(t, "foo.txt,1, 2 ,3", "foo.txt", 1, 2, 3)

	// spaces in fuzziness
	testParseIndexGood(t, "foo.txt,1,2, 3", "foo.txt", 1, 2, 3)
	testParseIndexGood(t, "foo.txt,1,2,3 ", "foo.txt", 1, 2, 3)
	testParseIndexGood(t, "foo.txt,1,2, 3 ", "foo.txt", 1, 2, 3)

	// comas in filename
	testParseIndexGood(t, "bar,foo.txt,0,0,0", "bar,foo.txt", 0, 0, 0)
	testParseIndexGood(t, "foo,bar,foo.txt,0,0,0", "foo,bar,foo.txt", 0, 0, 0)
}

// "bad" cases
func TestParseIndexBad(t *testing.T) {
	testParseIndexBad(t, "foo.txt,0,0", "invalid number of fields in")

	testParseIndexBad(t, "foo.txt,0.0,0,0", "failed to parse offset")
	testParseIndexBad(t, "foo.txt,a,0,0", "failed to parse offset")
	testParseIndexBad(t, "foo.txt,1a,0,0", "failed to parse offset")

	testParseIndexBad(t, "foo.txt,0,0.0,0", "failed to parse length")
	testParseIndexBad(t, "foo.txt,0,b,0", "failed to parse length")
	testParseIndexBad(t, "foo.txt,0,1b,0", "failed to parse length")
	testParseIndexBad(t, "foo.txt,0,66666,0", "failed to parse length") // out of 16 bits

	testParseIndexBad(t, "foo.txt,0,0,0.0", "failed to parse fuzziness")
	testParseIndexBad(t, "foo.txt,0,0,c", "failed to parse fuzziness")
	testParseIndexBad(t, "foo.txt,0,0,1c", "failed to parse fuzziness")
	testParseIndexBad(t, "foo.txt,0,0,256", "failed to parse fuzziness") // out of 8 bits
}
