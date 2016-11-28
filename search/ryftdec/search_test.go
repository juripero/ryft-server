package ryftdec

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/testfake"
	"github.com/stretchr/testify/assert"
)

// Check simple search results.
func TestEngineSearchBypass(t *testing.T) {
	SetLogLevelString(testLogLevel)
	taskId = 0 // reset to check intermediate file names

	f1 := newFake(1000, 10)
	f1.HomeDir = "/ryft-test"
	f1.HostName = "host"

	os.MkdirAll(filepath.Join(f1.MountPoint, f1.HomeDir, f1.Instance), 0755)
	ioutil.WriteFile(filepath.Join(f1.MountPoint, f1.HomeDir, "1.txt"), []byte(`
11111-hello-11111
22222-hello-22222
33333-hello-33333
44444-hello-44444
55555-hello-55555
`), 0644)
	ioutil.WriteFile(filepath.Join(f1.MountPoint, f1.HomeDir, "2.txt"), []byte{}, 0644)
	os.Mkdir(filepath.Join(f1.MountPoint, f1.HomeDir, "3.txt"), 0755)
	defer os.RemoveAll(filepath.Join(f1.MountPoint, f1.HomeDir))

	// valid (usual case)
	engine, err := NewEngine(f1, -1, false)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "*.txt")
		cfg.Width = 3
		cfg.ReportIndex = true
		cfg.ReportData = true

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			records, errors := testfake.Drain(res)

			// convert records to strings and sort
			strRecords := make([]string, 0, len(records))
			for _, rec := range records {
				strRecords = append(strRecords, rec.String())
			}
			sort.Strings(strRecords)

			assert.Empty(t, errors)
			assert.EqualValues(t, []string{
				`Record{{/tmp/ryft-test/1.txt#22, len:11, d:0}, data:"22-hello-22"}`,
				`Record{{/tmp/ryft-test/1.txt#4, len:11, d:0}, data:"11-hello-11"}`,
				`Record{{/tmp/ryft-test/1.txt#40, len:11, d:0}, data:"33-hello-33"}`,
				`Record{{/tmp/ryft-test/1.txt#58, len:11, d:0}, data:"44-hello-44"}`,
				`Record{{/tmp/ryft-test/1.txt#76, len:11, d:0}, data:"55-hello-55"}`,
			}, strRecords)

			if assert.EqualValues(t, 1, len(f1.SearchCfgLogTrace)) {
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["*.txt"], mode:"", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:"", keep-index:"", delim:"", index:true, data:true}`, f1.SearchCfgLogTrace[0].String())
			}
		}
	}
}

// check for simple AND
func TestEngineSearchAnd3(t *testing.T) {
	SetLogLevelString(testLogLevel)
	taskId = 0 // reset to check intermediate file names

	f1 := newFake(1000, 10)
	f1.HostName = "host-1"

	os.MkdirAll(filepath.Join(f1.MountPoint, f1.HomeDir, f1.Instance), 0755)
	ioutil.WriteFile(filepath.Join(f1.MountPoint, f1.HomeDir, "1.txt"), []byte(`
11111-hello-11111
22222-hello-22222
33333-hello-33333
44444-hello-44444
55555-hello-55555
`), 0644)
	defer os.RemoveAll(filepath.Join(f1.MountPoint, f1.HomeDir))

	// valid (usual case)
	engine, err := NewEngine(f1, -1, false)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello AND hell AND he", "1.txt")
		cfg.Width = 3
		cfg.ReportIndex = true
		cfg.ReportData = true

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			records, errors := testfake.Drain(res)

			// convert records to strings and sort
			strRecords := make([]string, 0, len(records))
			for _, rec := range records {
				strRecords = append(strRecords, rec.String())
			}
			sort.Strings(strRecords)

			assert.Empty(t, errors)
			assert.EqualValues(t, []string{
				`Record{{1.txt#22, len:8, d:0}, data:"22-hello"}`,
				`Record{{1.txt#4, len:8, d:0}, data:"11-hello"}`,
				`Record{{1.txt#40, len:8, d:0}, data:"33-hello"}`,
				`Record{{1.txt#58, len:8, d:0}, data:"44-hello"}`,
				`Record{{1.txt#76, len:8, d:0}, data:"55-hello"}`,
			}, strRecords)

			if assert.EqualValues(t, 3, len(f1.SearchCfgLogTrace)) {
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["1.txt"], mode:"", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-2.txt", keep-index:".work/.temp-idx-dec-00000001-2.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[0].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hell", WIDTH="3")), files:[".work/.temp-dat-dec-00000001-2.txt"], mode:"", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-3.txt", keep-index:".work/.temp-idx-dec-00000001-3.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[1].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("he", WIDTH="3")), files:[".work/.temp-dat-dec-00000001-3.txt"], mode:"", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-4.txt", keep-index:".work/.temp-idx-dec-00000001-4.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[2].String())
			}
		}
	}
}

// check for simple OR
func TestEngineSearchOr3(t *testing.T) {
	SetLogLevelString(testLogLevel)
	taskId = 0 // reset to check intermediate file names

	f1 := newFake(1000, 10)
	f1.HostName = "host-1"

	os.MkdirAll(filepath.Join(f1.MountPoint, f1.HomeDir, f1.Instance), 0755)
	ioutil.WriteFile(filepath.Join(f1.MountPoint, f1.HomeDir, "1.txt"), []byte(`
11111-hello-11111
22222-hello-22222
33333-hello-33333
44444-hello-44444
55555-hello-55555
`), 0644)
	defer os.RemoveAll(filepath.Join(f1.MountPoint, f1.HomeDir))

	// valid (usual case)
	engine, err := NewEngine(f1, -1, false)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello OR hell OR he", "1.txt")
		cfg.Width = 3
		cfg.ReportIndex = true
		cfg.ReportData = true

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			records, errors := testfake.Drain(res)

			// convert records to strings and sort
			strRecords := make([]string, 0, len(records))
			for _, rec := range records {
				strRecords = append(strRecords, rec.String())
			}
			sort.Strings(strRecords)

			assert.Empty(t, errors)
			assert.EqualValues(t, []string{
				`Record{{1.txt#22, len:10, d:0}, data:"22-hello-2"}`,
				`Record{{1.txt#22, len:11, d:0}, data:"22-hello-22"}`,
				`Record{{1.txt#22, len:8, d:0}, data:"22-hello"}`,
				`Record{{1.txt#4, len:10, d:0}, data:"11-hello-1"}`,
				`Record{{1.txt#4, len:11, d:0}, data:"11-hello-11"}`,
				`Record{{1.txt#4, len:8, d:0}, data:"11-hello"}`,
				`Record{{1.txt#40, len:10, d:0}, data:"33-hello-3"}`,
				`Record{{1.txt#40, len:11, d:0}, data:"33-hello-33"}`,
				`Record{{1.txt#40, len:8, d:0}, data:"33-hello"}`,
				`Record{{1.txt#58, len:10, d:0}, data:"44-hello-4"}`,
				`Record{{1.txt#58, len:11, d:0}, data:"44-hello-44"}`,
				`Record{{1.txt#58, len:8, d:0}, data:"44-hello"}`,
				`Record{{1.txt#76, len:10, d:0}, data:"55-hello-5"}`,
				`Record{{1.txt#76, len:11, d:0}, data:"55-hello-55"}`,
				`Record{{1.txt#76, len:8, d:0}, data:"55-hello"}`,
			}, strRecords)

			if assert.EqualValues(t, 3, len(f1.SearchCfgLogTrace)) {
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["1.txt"], mode:"", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-2.txt", keep-index:".work/.temp-idx-dec-00000001-2.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[0].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hell", WIDTH="3")), files:["1.txt"], mode:"", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-3.txt", keep-index:".work/.temp-idx-dec-00000001-3.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[1].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("he", WIDTH="3")), files:["1.txt"], mode:"", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-4.txt", keep-index:".work/.temp-idx-dec-00000001-4.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[2].String())
			}
		}
	}
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
	bad([]string{"a.txt", "b.dat"}, "", "unable to detect extension")
	bad([]string{"a.txt", "b.dat"}, "c.jpeg", "unable to detect extension")
	check([]string{}, "", "")
	check([]string{"foo/a.txt", "my.test/b.txt"}, "", ".txt")
	check([]string{"foo/a.txt", "my.test/b.txt"}, "data.txt", ".txt")
	check([]string{"foo/*.txt", "my.test/*txt"}, "", ".txt")
	check([]string{"foo/*.txt", "my.test/*"}, "data.txt", ".txt")
	check([]string{"my.test/*"}, "data.txt", ".txt")
	check([]string{"nyctaxi/xml/2015/yellow/*"}, "ryftnyctest.nxml", ".nxml")
}
