package ryftdec

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/testfake"
	"github.com/getryft/ryft-server/search/utils/catalog"
	"github.com/stretchr/testify/assert"
)

// Check simple search results.
func TestEngineSearchBypass(t *testing.T) {
	testSetLogLevel()
	taskId = 0 // reset to check intermediate file names

	f1 := testNewFake()
	f1.HomeDir = "/ryft-test"
	f1.HostName = "host"

	assert.NoError(t, os.RemoveAll(filepath.Join(f1.MountPoint, f1.HomeDir)))
	defer os.RemoveAll(filepath.Join(f1.MountPoint, f1.HomeDir))
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

	// valid (usual case)
	engine, err := NewEngine(f1, nil)
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
				// NOTE, files:["1.txt"] - since it is expanded!
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["1.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:"", keep-index:"", delim:"", index:true, data:true}`, f1.SearchCfgLogTrace[0].String())
			}
		}
	}
}

// check for simple AND
func TestEngineSearchAnd3(t *testing.T) {
	testSetLogLevel()
	taskId = 0 // reset to check intermediate file names

	f1 := testNewFake()
	f1.HostName = "host-1"

	assert.NoError(t, os.RemoveAll(filepath.Join(f1.MountPoint, f1.HomeDir)))
	defer os.RemoveAll(filepath.Join(f1.MountPoint, f1.HomeDir))
	os.MkdirAll(filepath.Join(f1.MountPoint, f1.HomeDir, f1.Instance), 0755)
	ioutil.WriteFile(filepath.Join(f1.MountPoint, f1.HomeDir, "1.txt"), []byte(`
11111-hello-11111
22222-hello-22222
33333-hello-33333
44444-hello-44444
55555-hello-55555
`), 0644)

	// valid (usual case)
	engine, err := NewEngine(f1, nil)
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
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["1.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-2.txt", keep-index:".work/.temp-idx-dec-00000001-2.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[0].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hell", WIDTH="3")), files:[".work/.temp-dat-dec-00000001-2.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-3.txt", keep-index:".work/.temp-idx-dec-00000001-3.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[1].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("he", WIDTH="3")), files:[".work/.temp-dat-dec-00000001-3.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-4.txt", keep-index:".work/.temp-idx-dec-00000001-4.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[2].String())
			}
		}
	}
}

// check for simple OR
func TestEngineSearchOr3(t *testing.T) {
	testSetLogLevel()

	f1 := testNewFake()
	f1.HostName = "host-1"

	assert.NoError(t, os.RemoveAll(filepath.Join(f1.MountPoint, f1.HomeDir)))
	defer os.RemoveAll(filepath.Join(f1.MountPoint, f1.HomeDir))
	os.MkdirAll(filepath.Join(f1.MountPoint, f1.HomeDir, f1.Instance), 0755)
	ioutil.WriteFile(filepath.Join(f1.MountPoint, f1.HomeDir, "1.txt"), []byte(`
11111-hello-11111
22222-hello-22222
33333-hello-33333
44444-hello-44444
55555-hello-55555
`), 0644)

	// valid (usual case)
	engine, err := NewEngine(f1, nil)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("{hello} OR {hell} OR {he}", "1.txt")
		cfg.Width = 3
		cfg.ReportIndex = true
		cfg.ReportData = true

		taskId = 0 // reset to check intermediate file names
		f1.SearchCfgLogTrace = nil

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
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["1.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-2.txt", keep-index:".work/.temp-idx-dec-00000001-2.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[0].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hell", WIDTH="3")), files:["1.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-3.txt", keep-index:".work/.temp-idx-dec-00000001-3.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[1].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("he", WIDTH="3")), files:["1.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-4.txt", keep-index:".work/.temp-idx-dec-00000001-4.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[2].String())
			}
		}
	}

	// remove duplicates (usual case)
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("{hello} OR {hello} OR {hello}", "1.txt")
		cfg.Width = 3
		cfg.ReportIndex = true
		cfg.ReportData = true

		taskId = 0 // reset to check intermediate file names
		f1.SearchCfgLogTrace = nil

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
				`Record{{1.txt#22, len:11, d:0}, data:"22-hello-22"}`,
				`Record{{1.txt#4, len:11, d:0}, data:"11-hello-11"}`,
				`Record{{1.txt#40, len:11, d:0}, data:"33-hello-33"}`,
				`Record{{1.txt#58, len:11, d:0}, data:"44-hello-44"}`,
				`Record{{1.txt#76, len:11, d:0}, data:"55-hello-55"}`,
			}, strRecords)

			if assert.EqualValues(t, 3, len(f1.SearchCfgLogTrace)) {
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["1.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-2.txt", keep-index:".work/.temp-idx-dec-00000001-2.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[0].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["1.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-3.txt", keep-index:".work/.temp-idx-dec-00000001-3.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[1].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["1.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-4.txt", keep-index:".work/.temp-idx-dec-00000001-4.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[2].String())
			}
		}
	}
}

// add a part to catalog
func testAddToCatalog(cat *catalog.Catalog, filename string, offset int64, data string) error {
	dataPath, dataPos, delim, err := cat.AddFilePart(filename, offset, int64(len(data)), nil)
	if err != nil {
		return fmt.Errorf("failed to add file part: %s", err)
	}

	dataDir, _ := filepath.Split(dataPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directories: %s", err)
	}

	// write file content
	f, err := os.OpenFile(dataPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open data file: %s", err)
	}
	defer f.Close()

	_, err = f.Seek(dataPos, os.SEEK_SET /*TODO: io.SeekStart*/)
	if err != nil {
		return fmt.Errorf("failed to seek data file: %s", err)
	}

	n, err := f.WriteString(data)
	if err != nil {
		return fmt.Errorf("failed to copy data: %s", err)
	}
	if n != len(data) {
		return fmt.Errorf("only %d bytes copied of %d", n, len(data))
	}

	// write data delimiter
	if len(delim) > 0 {
		nn, err := f.WriteString(delim)
		if err != nil {
			return fmt.Errorf("failed to write delimiter: %s", err)
		}
		if nn != len(delim) {
			return fmt.Errorf("only %d bytes copied of %d", nn, len(delim))
		}
	}

	return nil // OK
}

// Check catalog search results.
func TestEngineSearchCatalog(t *testing.T) {
	testSetLogLevel()

	f1 := testNewFake()
	f1.HomeDir = "/ryft-test1"
	f1.HostName = "host"

	catalog.DefaultDataDelimiter = "\n"
	assert.NoError(t, os.RemoveAll(filepath.Join(f1.MountPoint, f1.HomeDir)))
	defer os.RemoveAll(filepath.Join(f1.MountPoint, f1.HomeDir))
	os.MkdirAll(filepath.Join(f1.MountPoint, f1.HomeDir, f1.Instance), 0755)
	cat, err := catalog.OpenCatalogNoCache(filepath.Join(f1.MountPoint, f1.HomeDir, "cat.txt"))
	if !assert.NoError(t, err) {
		return
	}
	defer cat.Close()

	// part-1
	err = testAddToCatalog(cat, "1.txt", -1, `hello-00000`)
	if !assert.NoError(t, err) {
		return
	}

	// part-2
	err = testAddToCatalog(cat, "2.txt", -1, `11111-hello-11111
22222-hello-22222
33333-hello-33333
44444-hello-44444
55555-hello-55555`)
	if !assert.NoError(t, err) {
		return
	}

	// part-3
	err = testAddToCatalog(cat, "3.txt", -1, `99999-hello`)
	if !assert.NoError(t, err) {
		return
	}

	// valid (width=0)
	engine, err := NewEngine(f1, nil)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "*.txt")
		cfg.Width = 0
		cfg.ReportIndex = true
		cfg.ReportData = true

		taskId = 0 // reset to check intermediate file names
		f1.SearchCfgLogTrace = nil

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
				`Record{{1.txt#0, len:5, d:0}, data:"hello"}`,
				`Record{{2.txt#24, len:5, d:0}, data:"hello"}`,
				`Record{{2.txt#42, len:5, d:0}, data:"hello"}`,
				`Record{{2.txt#6, len:5, d:0}, data:"hello"}`,
				`Record{{2.txt#60, len:5, d:0}, data:"hello"}`,
				`Record{{2.txt#78, len:5, d:0}, data:"hello"}`,
				`Record{{3.txt#6, len:5, d:0}, data:"hello"}`,
			}, strRecords)

			if assert.EqualValues(t, 1, len(f1.SearchCfgLogTrace)) {
				f1.SearchCfgLogTrace[0].Files = []string{"*.txt"} // skip catalog's data files
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello")), files:["*.txt"], mode:"g", width:0, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-1", keep-index:".work/.temp-idx-dec-00000001-1.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[0].String())
			}
		}
	}

	// valid (width=3)
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "*.txt")
		cfg.Width = 3
		cfg.ReportIndex = true
		cfg.ReportData = true

		taskId = 0 // reset to check intermediate file names
		f1.SearchCfgLogTrace = nil

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
				`Record{{1.txt#0, len:8, d:0}, data:"hello-00"}`,
				`Record{{2.txt#21, len:11, d:0}, data:"22-hello-22"}`,
				`Record{{2.txt#3, len:11, d:0}, data:"11-hello-11"}`,
				`Record{{2.txt#39, len:11, d:0}, data:"33-hello-33"}`,
				`Record{{2.txt#57, len:11, d:0}, data:"44-hello-44"}`,
				`Record{{2.txt#75, len:11, d:0}, data:"55-hello-55"}`,
				`Record{{3.txt#3, len:8, d:0}, data:"99-hello"}`,
			}, strRecords)

			if assert.EqualValues(t, 1, len(f1.SearchCfgLogTrace)) {
				f1.SearchCfgLogTrace[0].Files = []string{"*.txt"} // skip catalog's data files
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["*.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-1", keep-index:".work/.temp-idx-dec-00000001-1.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[0].String())
			}
		}
	}

	// valid (width=line)
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "*.txt")
		cfg.Width = -1 // line
		cfg.ReportIndex = true
		cfg.ReportData = true

		taskId = 0 // reset to check intermediate file names
		f1.SearchCfgLogTrace = nil

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
				`Record{{1.txt#0, len:11, d:0}, data:"hello-00000"}`,
				`Record{{2.txt#0, len:17, d:0}, data:"11111-hello-11111"}`,
				`Record{{2.txt#18, len:17, d:0}, data:"22222-hello-22222"}`,
				`Record{{2.txt#36, len:17, d:0}, data:"33333-hello-33333"}`,
				`Record{{2.txt#54, len:17, d:0}, data:"44444-hello-44444"}`,
				`Record{{2.txt#72, len:17, d:0}, data:"55555-hello-55555"}`,
				`Record{{3.txt#0, len:11, d:0}, data:"99999-hello"}`,
			}, strRecords)

			if assert.EqualValues(t, 1, len(f1.SearchCfgLogTrace)) {
				f1.SearchCfgLogTrace[0].Files = []string{"*.txt"} // skip catalog's data files
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", LINE="true")), files:["*.txt"], mode:"g", width:-1, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-1", keep-index:".work/.temp-idx-dec-00000001-1.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[0].String())
			}
		}
	}

	// filter (width=3)
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig(`RAW_TEXT CONTAINS EXACT("hello", FILTER="[1|3]\.txt")`, "*.txt")
		cfg.Width = 3
		cfg.ReportIndex = true
		cfg.ReportData = true

		taskId = 0 // reset to check intermediate file names
		f1.SearchCfgLogTrace = nil

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
				`Record{{1.txt#0, len:8, d:0}, data:"hello-00"}`,
				//`Record{{2.txt#21, len:11, d:0}, data:"22-hello-22"}`,
				//`Record{{2.txt#3, len:11, d:0}, data:"11-hello-11"}`,
				//`Record{{2.txt#39, len:11, d:0}, data:"33-hello-33"}`,
				//`Record{{2.txt#57, len:11, d:0}, data:"44-hello-44"}`,
				//`Record{{2.txt#75, len:11, d:0}, data:"55-hello-55"}`,
				`Record{{3.txt#3, len:8, d:0}, data:"99-hello"}`,
			}, strRecords)

			if assert.EqualValues(t, 1, len(f1.SearchCfgLogTrace)) {
				f1.SearchCfgLogTrace[0].Files = []string{"*.txt"} // skip catalog's data files
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["*.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-1", keep-index:".work/.temp-idx-dec-00000001-1.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[0].String())
			}
		}
	}
}

// check bad cases
func TestEngineSearchBad(t *testing.T) {
	testSetLogLevel()

	f1 := testNewFake()
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
	ioutil.WriteFile(filepath.Join(f1.MountPoint, f1.HomeDir, "1.dat"), []byte(`
11111-hello-11111
22222-hello-22222
33333-hello-33333
44444-hello-44444
55555-hello-55555
`), 0644)
	defer os.RemoveAll(filepath.Join(f1.MountPoint, f1.HomeDir))

	engine, err := NewEngine(f1, nil)
	if !assert.NoError(t, err) {
		return
	}

	// DATA without INDEX
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "*.txt")
		cfg.ReportIndex = false
		cfg.ReportData = true

		_, err := engine.Search(cfg)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to report DATA without INDEX")
		}
	}

	// .. in file name
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "../*.txt")
		cfg.ReportIndex = false
		cfg.ReportData = false

		_, err := engine.Search(cfg)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "is not relative to home")
		}
	}

	// no file found
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "*.missing")
		cfg.ReportIndex = false
		cfg.ReportData = false

		_, err := engine.Search(cfg)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "no any valid file or catalog found")
		}
	}

	// bad extension
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello AND bye", "*.txt", "*.dat")
		cfg.ReportIndex = true
		cfg.ReportData = false

		_, err := engine.Search(cfg)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to detect extension")
		}
	}

	// bad query
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("()", "*.txt")
		cfg.ReportIndex = true
		cfg.ReportData = false

		_, err := engine.Search(cfg)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "failed to decompose query")
		}
	}

	// failed to do search
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("{hello} AND {no}", "*.txt")
		cfg.ReportIndex = false
		cfg.ReportData = false
		f1.SearchReportError = fmt.Errorf("stop-by-test")

		engine.CompatMode = true
		res, err := engine.Search(cfg)
		if err == nil {
			_, errors := testfake.Drain(res)
			if len(errors) > 0 {
				err = errors[0] // get first one
			}
		}
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "stop-by-test")
		}

		f1.SearchReportError = nil
	}

	// failed to do search
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("{hello} XOR {no}", "*.txt")
		cfg.ReportIndex = false
		cfg.ReportData = false

		engine.CompatMode = true
		res, err := engine.Search(cfg)
		if err == nil {
			_, errors := testfake.Drain(res)
			if len(errors) > 0 {
				err = errors[0] // get first one
			}
		}
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "XOR is not implemented yet")
		}
	}
}

// check for [] AND ()
func TestEngineSearchIndexAnd(t *testing.T) {
	testSetLogLevel()

	f1 := testNewFake()
	f1.HomeDir = "/ryft-test2"
	f1.HostName = "host"

	catalog.DefaultDataDelimiter = "\n"
	assert.NoError(t, os.RemoveAll(filepath.Join(f1.MountPoint, f1.HomeDir)))
	defer os.RemoveAll(filepath.Join(f1.MountPoint, f1.HomeDir))
	os.MkdirAll(filepath.Join(f1.MountPoint, f1.HomeDir, f1.Instance), 0755)
	cat, err := catalog.OpenCatalogNoCache(filepath.Join(f1.MountPoint, f1.HomeDir, "cat.txt"))
	if !assert.NoError(t, err) {
		return
	}
	defer cat.Close()

	// part-1
	err = testAddToCatalog(cat, "1.txt", -1, `hello-00000`)
	if !assert.NoError(t, err) {
		return
	}

	// part-2
	err = testAddToCatalog(cat, "X.dat", -1, `11111-hello-11111
22222-hello-22222
33333-hello-33333
44444-hello-44444
55555-hello-55555`)
	if !assert.NoError(t, err) {
		return
	}

	// part-3
	err = testAddToCatalog(cat, "3.txt", -1, `99999-hello`)
	if !assert.NoError(t, err) {
		return
	}

	ioutil.WriteFile(filepath.Join(f1.MountPoint, f1.HomeDir, "X.dat"), []byte(`
aaaaa-hello-aaaaa
bbbbb-hello-bbbbb
ccccc-hello-ccccc
ddddd-hello-ddddd
eeeee-hello-eeeee
`), 0644)

	// valid (width=0)
	engine, err := NewEngine(f1, nil)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("[hello] AND (hello)", "cat.txt")
		cfg.Width = 3
		cfg.ReportIndex = true
		cfg.ReportData = true

		taskId = 0 // reset to check intermediate file names
		f1.SearchCfgLogTrace = nil

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
				`Record{{X.dat#22, len:11, d:0}, data:"bb-hello-bb"}`,
				`Record{{X.dat#4, len:11, d:0}, data:"aa-hello-aa"}`,
				`Record{{X.dat#40, len:11, d:0}, data:"cc-hello-cc"}`,
				`Record{{X.dat#58, len:11, d:0}, data:"dd-hello-dd"}`,
				`Record{{X.dat#76, len:11, d:0}, data:"ee-hello-ee"}`,
			}, strRecords)

			if assert.EqualValues(t, 2, len(f1.SearchCfgLogTrace)) {
				f1.SearchCfgLogTrace[0].Files = []string{"*.txt"} // skip catalog's data files
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["*.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-2.txt", keep-index:".work/.temp-idx-dec-00000001-2.txt", delim:"\n", index:false, data:false}`, f1.SearchCfgLogTrace[0].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["X.dat"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-3.txt", keep-index:".work/.temp-idx-dec-00000001-3.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[1].String())
			}
		}
	}

	ioutil.WriteFile(filepath.Join(f1.MountPoint, f1.HomeDir, "1.txt"), []byte(`
xxxxx-hello-xxxxx
`), 0644)

	ioutil.WriteFile(filepath.Join(f1.MountPoint, f1.HomeDir, "3.txt"), []byte(`
yyyyy-hello-yyyyy
zzzzz-hello-zzzzz
`), 0644)

	// use file filter
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig(`[RAW_TEXT CONTAINS EXACT("hello", FF="[0-9]\.txt")] AND (hello)`, "cat.txt")
		cfg.Width = 3
		cfg.ReportIndex = true
		cfg.ReportData = true

		taskId = 0 // reset to check intermediate file names
		f1.SearchCfgLogTrace = nil

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
				`Record{{1.txt#4, len:11, d:0}, data:"xx-hello-xx"}`,
				`Record{{3.txt#22, len:11, d:0}, data:"zz-hello-zz"}`,
				`Record{{3.txt#4, len:11, d:0}, data:"yy-hello-yy"}`,
			}, strRecords)

			if assert.EqualValues(t, 2, len(f1.SearchCfgLogTrace)) {
				f1.SearchCfgLogTrace[0].Files = []string{"*.txt"} // skip catalog's data files
				sort.Strings(f1.SearchCfgLogTrace[1].Files)
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["*.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-2.txt", keep-index:".work/.temp-idx-dec-00000001-2.txt", delim:"\n", index:false, data:false}`, f1.SearchCfgLogTrace[0].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["1.txt" "3.txt"], mode:"g", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-3.txt", keep-index:".work/.temp-idx-dec-00000001-3.txt", delim:"", index:false, data:false}`, f1.SearchCfgLogTrace[1].String())
			}
		}
	}
}

// check for parse final results
func TestRyftFinalResults(t *testing.T) {
	testSetLogLevel()

	home := "/tmp/ryft-test3"

	check := func(dataPath, indexPath []string, width []int, expectedData, expectedIndex string) {
		n := len(dataPath)
		if len(indexPath) < n {
			n = len(indexPath)
		}
		if len(width) < n {
			n = len(width)
		}

		task := NewTask(search.NewEmptyConfig())
		mpp, _ := NewInMemoryPostProcessing()
		ryftCalls := make([]RyftCall, n)
		for i := 0; i < n; i++ {
			opt := uint32(0)
			if i+1 == n {
				opt = 1 // final
			}
			rc := RyftCall{
				DataFile:  dataPath[i],
				IndexFile: indexPath[i],
				Delimiter: "\n",
				Width:     width[i],
			}
			ryftCalls[i] = rc

			err := mpp.AddRyftResults(filepath.Join(home, rc.DataFile),
				filepath.Join(home, rc.IndexFile),
				rc.Delimiter, rc.Width, opt)
			if !assert.NoError(t, err) {
				return
			}
		}

		mux := search.NewResult()
		_, err := mpp.DrainFinalResults(task, mux,
			"data.out", "index.out", "\n", "",
			home, ryftCalls, "")
		if !assert.NoError(t, err) {
			return
		}

		if data, err := ioutil.ReadFile(filepath.Join(home, "data.out")); assert.NoError(t, err) {
			assert.EqualValues(t, expectedData, string(data))
		}
		if data, err := ioutil.ReadFile(filepath.Join(home, "index.out")); assert.NoError(t, err) {
			assert.EqualValues(t, expectedIndex, string(data))
		}
	}

	catalog.DefaultDataDelimiter = "\n"
	assert.NoError(t, os.RemoveAll(home))
	defer os.RemoveAll(home)
	os.MkdirAll(home, 0755)

	ioutil.WriteFile(filepath.Join(home, "X.dat"),
		[]byte(`aaaaa-hello-aaaaa
bbbbb-hello-bbbbb
ccccc-hello-ccccc
ddddd-hello-ddddd
eeeee-hello-eeeee
`), 0644)

	// "hello", W=3
	ioutil.WriteFile(filepath.Join(home, "data-0.dat"),
		[]byte(`bb-hello-bb
aa-hello-aa
dd-hello-dd
cc-hello-cc
ee-hello-ee
`), 0644)
	ioutil.WriteFile(filepath.Join(home, "index-0.txt"), // shuffle output
		[]byte(`/tmp/ryft-test3/X.dat,21,11,0
/tmp/ryft-test3/X.dat,3,11,0
/tmp/ryft-test3/X.dat,57,11,0
/tmp/ryft-test3/X.dat,39,11,0
/tmp/ryft-test3/X.dat,75,11,0`), 0644)

	check([]string{"data-0.dat"},
		[]string{"index-0.txt"},
		[]int{3},
		`bb-hello-bb
aa-hello-aa
dd-hello-dd
cc-hello-cc
ee-hello-ee
`,
		`X.dat,21,11,0
X.dat,3,11,0
X.dat,57,11,0
X.dat,39,11,0
X.dat,75,11,0
`)

	// "hello", W=3 AND "hello", W=5
	ioutil.WriteFile(filepath.Join(home, "data-1a.dat"),
		[]byte(`bb-hello-bb
aa-hello-aa
dd-hello-dd
cc-hello-cc
ee-hello-ee
`), 0644)
	ioutil.WriteFile(filepath.Join(home, "index-1a.txt"), // shuffle output
		[]byte(`/tmp/ryft-test3/X.dat,21,11,0
/tmp/ryft-test3/X.dat,3,11,0
/tmp/ryft-test3/X.dat,57,11,0
/tmp/ryft-test3/X.dat,39,11,0
/tmp/ryft-test3/X.dat,75,11,0`), 0644)
	ioutil.WriteFile(filepath.Join(home, "data-1b.dat"),
		[]byte(`bb-hello-bb
a
b
aa-hello-aa
d
b
dd-hello-dd
c
d
cc-hello-cc
e
c
ee-hello-ee
`), 0644)
	ioutil.WriteFile(filepath.Join(home, "index-1b.txt"),
		[]byte(`/tmp/ryft-test3/data-1a.dat,0,13,0
/tmp/ryft-test3/data-1a.dat,10,15,0
/tmp/ryft-test3/data-1a.dat,22,15,0
/tmp/ryft-test3/data-1a.dat,34,15,0
/tmp/ryft-test3/data-1a.dat,46,13,0`), 0644)

	check([]string{"data-1a.dat", "data-1b.dat"},
		[]string{"index-1a.txt", "index-1b.txt"},
		[]int{3, 5},
		`bb-hello-bb
aa-hello-aa
dd-hello-dd
cc-hello-cc
ee-hello-ee
`,
		`X.dat,21,11,0
X.dat,3,11,0
X.dat,57,11,0
X.dat,39,11,0
X.dat,75,11,0
`)
}
