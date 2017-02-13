package ryftdec

import (
	// "fmt"
	"testing"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils/query"
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

// ryftcall test
func TestRyftCall(t *testing.T) {
	rc := RyftCall{
		DataFile:  "1.dat",
		IndexFile: "1.txt",
		Delimiter: "\n",
		Width:     3,
	}

	assert.EqualValues(t, `RyftCall{data:1.dat, index:1.txt, delim:#0a, width:3}`, rc.String())
}

// search result test
func TestSearchResult(t *testing.T) {
	var res SearchResult

	// test empty results
	assert.Nil(t, res.Stat)
	assert.Empty(t, res.GetDataFiles())
	assert.EqualValues(t, 0, res.Matches())
}

// combine stat test
func TestCombineStat(t *testing.T) {
	mux := search.NewStat("h1")
	s1 := search.NewStat("s1")
	s1.Matches = 1
	s1.TotalBytes = 1024 * 1024

	combineStat(mux, s1)
	assert.EqualValues(t, mux.Matches, s1.Matches)
	assert.EqualValues(t, mux.TotalBytes, s1.TotalBytes)
	assert.EqualValues(t, mux.FabricDuration, s1.FabricDuration)
	assert.EqualValues(t, mux.Duration, s1.Duration)
	assert.InDelta(t, 0.0, mux.FabricDataRate, 1e-5)
	assert.InDelta(t, 0.0, mux.DataRate, 1e-5)

	s1.Duration = 2000
	s1.FabricDuration = 1000
	combineStat(mux, s1)
	assert.EqualValues(t, mux.Matches, 2*s1.Matches)
	assert.EqualValues(t, mux.TotalBytes, 2*s1.TotalBytes)
	assert.EqualValues(t, mux.FabricDuration, s1.FabricDuration)
	assert.EqualValues(t, mux.Duration, s1.Duration)
	assert.InDelta(t, 2.0, mux.FabricDataRate, 1e-5)
	assert.InDelta(t, 1.0, mux.DataRate, 1e-5)
}

// find file filter test
func TestFindFilter(t *testing.T) {
	q := query.Query{
		Operator: "-",
		Arguments: []query.Query{
			{Simple: &query.SimpleQuery{
				Options: query.Options{FileFilter: "A"},
			}},
			{Simple: &query.SimpleQuery{
				Options: query.Options{FileFilter: "-"},
			}},
			{Simple: &query.SimpleQuery{
				Options: query.Options{FileFilter: "B"},
			}},
		},
	}

	assert.EqualValues(t, "A", findFirstFilter(q))
	assert.EqualValues(t, "B", findLastFilter(q))
}

// test call script
func TestCallScript(t *testing.T) {
	// good case
	check := func(script []string, in string, expectedOut string) {
		out, err := callScript(script, []byte(in))
		if assert.NotNil(t, out) && assert.NoError(t, err) {
			assert.EqualValues(t, expectedOut, out)
		}
	}

	// bad case
	bad := func(script []string, in string, expectedError string) {
		out, err := callScript(script, []byte(in))
		if assert.Nil(t, out) && assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	check([]string{"/bin/cat"}, "hello", "hello")
	check([]string{"/bin/grep", "hell"}, "apple\nhello\norange\n", "hello\n")
	bad([]string{"/bin/false"}, "hello", "skipped")
	bad([]string{}, "hello", "no script provided")
}
