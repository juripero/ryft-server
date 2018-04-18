/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2015, Ryft Systems, Inc.
 * All rights reserved.
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *   this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *   this list of conditions and the following disclaimer in the documentation and/or
 *   other materials provided with the distribution.
 * 3. All advertising materials mentioning features or use of this software must display the following acknowledgement:
 *   This product includes software developed by Ryft Systems, Inc.
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
 * THIS SOFTWARE IS PROVIDED BY RYFT SYSTEMS, INC. ''AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL RYFT SYSTEMS, INC. BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 * ============
 */

package ryftdec

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils/query"
	"github.com/stretchr/testify/assert"
)

// custom value to JSON
func asJson(v interface{}) string {
	d, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return string(d)
}

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

	assert.EqualValues(t, `RyftCall{data:1.dat, index:1.txt, delim:#0a, width:3, json-array:false}`, rc.String())
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

// detect file format
func TestFileFormat(t *testing.T) {
	SetLogLevelString(testLogLevel)

	engine := Engine{
		skipPatterns: []string{"*.txt"},
		jsonPatterns: []string{"*.json"},
		xmlPatterns:  []string{"*.xml", "foo/*.myxml"},
		csvPatterns:  []string{"*.csv", "foo/*.mycsv"},
	}

	check := func(path string, expected, expectedRoot string) {
		format, root, err := engine.detectFileFormat(path)
		if assert.NoError(t, err) {
			assert.EqualValues(t, expected, format)
			assert.EqualValues(t, expectedRoot, root)
		}
	}

	bad := func(path string, expected ...string) {
		_, _, err := engine.detectFileFormat(path)
		if assert.Error(t, err) {
			for _, msg := range expected {
				assert.Contains(t, err.Error(), msg)
			}
		}
	}
	_ = bad

	root := fmt.Sprintf("/tmp/ryft-%u", time.Now().UnixNano())
	assert.NoError(t, os.MkdirAll(root, 0755))
	defer os.RemoveAll(root)
	ioutil.WriteFile(filepath.Join(root, "1.xml"),
		[]byte(`<?xml version="1.0" encoding="UTF-8"?>
<root>
  <rec>
  </rec>
</root>`), 0644)
	ioutil.WriteFile(filepath.Join(root, "2.xmlx"),
		[]byte(`<?xml version="1.0" encoding="UTF-8"?>
<root>
  <rec>
  </rec>
</root>`), 0644)
	ioutil.WriteFile(filepath.Join(root, "1.csv"),
		[]byte(`1,2,3
4,5,6
`), 0644)
	ioutil.WriteFile(filepath.Join(root, "2.csvx"),
		[]byte(`1,2,3
4,5,6
`), 0644)
	ioutil.WriteFile(filepath.Join(root, "1.bin"),
		[]byte{0, 0, 0, 0, 0, 0, 0, 0, 0}, 0644)

	check(filepath.Join(root, "1.xml"), "XML", "rec") // by extension
	// check(filepath.Join(root, "3.xml"), "XML", "rec")  // by extension
	check(filepath.Join(root, "2.xmlx"), "XML", "rec") // by content

	// check(filepath.Join(root, "foo/3.myxml"), "XML", "rec") // by extension
	bad(filepath.Join(root, "3.myxml"), "no such file or directory")

	check(filepath.Join(root, "1.csv"), "CSV", "")  // by extension
	check(filepath.Join(root, "3.csv"), "CSV", "")  // by extension
	check(filepath.Join(root, "2.csvx"), "CSV", "") // by content

	check(filepath.Join(root, "foo/3.mycsv"), "CSV", "") // by extension
	bad(filepath.Join(root, "3.mycsv"), "no such file or directory")

	bad(filepath.Join(root, "foo/1.bin"), "no such file or directory")
	bad(filepath.Join(root, "1.bin"), "unknown file format")

	check(filepath.Join(root, "1.txt"), "", "")      // skip
	check(filepath.Join(root, "1.json"), "JSON", "") // by extension
}
