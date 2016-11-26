package ryftdec

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/stretchr/testify/assert"
)

// Run asynchronous "/search" or "/count" operation.
func (fe *fakeEngine) Search(cfg *search.Config) (*search.Result, error) {
	if fe.ErrorForSearch != nil {
		return nil, fe.ErrorForSearch
	}

	log.WithField("config", cfg).Infof("[fake]: start /search")
	cfgCopy := *cfg
	fe.searchDone = append(fe.searchDone, &cfgCopy)

	q, err := ParseQuery(cfg.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %s", err)
	}
	q = Optimize(q, -1)
	if q.Simple == nil {
		return nil, fmt.Errorf("no simple query parsed")
	}

	var text string
	if m := regexp.MustCompile(`\((RAW_TEXT|RECORD|RECORD\.\.*) (CONTAINS|NOT_CONTAINS|EQUALS|NOT_EQUALS) (EXACT|HAMMING|EDIT_DISTANCE)\("([^"]*)".*\)\)`).FindStringSubmatch(q.Simple.ExprNew); len(m) > 4 {
		text = m[4]
	}
	if len(text) == 0 {
		return nil, fmt.Errorf("nothing to search")
	}
	log.WithField("text", text).Infof("[fake]: text to search")

	res := search.NewResult()
	go func() {
		defer log.WithField("result", res).Infof("[fake]: done /search")

		defer res.Close()
		defer res.ReportDone()

		var datWr *bufio.Writer
		var idxWr *bufio.Writer

		res.Stat = search.NewStat(fe.Host)
		started := time.Now()
		defer func() {
			res.Stat.Duration = uint64(time.Since(started).Nanoseconds() / 1000)
			res.Stat.FabricDuration = res.Stat.Duration / 2
		}()

		for _, f := range cfg.Files {
			if res.IsCancelled() {
				break
			}

			mask := filepath.Join(fe.MountPoint, fe.HomeDir, f)
			matches, err := filepath.Glob(mask)
			//log.WithField("files", matches).WithField("mask", mask).Infof("[fake]: glob matches")
			if err != nil {
				res.ReportError(fmt.Errorf("failed to glob: %s", err))
				return // failed
			}
			for _, file := range matches {
				if res.IsCancelled() {
					break
				}

				if info, err := os.Stat(file); err != nil {
					res.ReportError(fmt.Errorf("failed to stat: %s", err))
					continue
				} else if info.IsDir() {
					continue // skip dirs
				} else if info.Size() == 0 {
					continue // skip empty files
				}

				f, err := os.Open(file)
				if err != nil {
					res.ReportError(fmt.Errorf("failed to open: %s", err))
					continue
				}
				defer f.Close()
				data, err := ioutil.ReadAll(f)
				if err != nil {
					res.ReportError(fmt.Errorf("failed to read: %s", err))
					continue
				}

				log.WithField("file", file).WithField("length", len(data)).Infof("[fake]: searching file")
				res.Stat.TotalBytes += uint64(len(data))
				for start, i := 0, 0; !res.IsCancelled() && i < 100; i++ {
					n := bytes.Index(data[start:], []byte(text))
					//log.WithField("start", start).WithField("found", n).Infof("[fake]: search iteration")
					if n < 0 {
						break
					}

					start += n + 1 // find next
					res.Stat.Matches++

					d_beg := start - 1 - int(cfg.Width)
					d_len := len(text) + 2*int(cfg.Width)
					if d_beg < 0 {
						d_len += d_beg
						d_beg = 0
					}
					if d_beg+d_len > len(data) {
						d_len = len(data) - d_beg
					}

					idx := search.NewIndex(file, uint64(d_beg), uint64(d_len))
					idx.UpdateHost(fe.Host)
					d := data[d_beg : d_beg+d_len]
					rec := search.NewRecord(idx, d)

					if idxWr == nil && len(cfg.KeepIndexAs) != 0 {
						f, err := os.Create(filepath.Join(fe.MountPoint, fe.HomeDir, cfg.KeepIndexAs))
						if err == nil {
							log.WithField("path", f.Name()).Infof("[fake]: saving INDEX to...")
							idxWr = bufio.NewWriter(f)
							defer f.Close()
							defer idxWr.Flush()
						} else {
							log.WithError(err).Errorf("[fake]: failed to save INDEX")
						}
					}
					if datWr == nil && len(cfg.KeepDataAs) != 0 {
						f, err := os.Create(filepath.Join(fe.MountPoint, fe.HomeDir, cfg.KeepDataAs))
						if err == nil {
							log.WithField("path", f.Name()).Infof("[fake]: saving DATA to...")
							datWr = bufio.NewWriter(f)
							defer f.Close()
							defer datWr.Flush()
						} else {
							log.WithError(err).Errorf("[fake]: failed to save DATA")
						}
					}
					if idxWr != nil {
						idxWr.WriteString(fmt.Sprintf("%s,%d,%d,%d\n", idx.File, idx.Offset, idx.Length, idx.Fuzziness))
					}
					if datWr != nil {
						datWr.Write(rec.RawData)
						datWr.WriteString(cfg.Delimiter)
					}

					if cfg.ReportIndex {
						res.ReportRecord(rec)
						log.WithField("rec", rec).Infof("[fake]: report record")
					} else {
						rec.Release()
					}
				}
			}
		}
	}()

	return res, nil // OK for now
}

// drain the results
func drain(res *search.Result) (records int, errors int) {
	for {
		select {
		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				errors++
			}

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				records++
			}

		case <-res.DoneChan:
			for _ = range res.ErrorChan {
				errors++
			}
			for _ = range res.RecordChan {
				records++
			}

			return
		}
	}
}

// drain the results
func drainFull(res *search.Result) (records []*search.Record, errors []error) {
	for {
		select {
		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				errors = append(errors, err)
			}

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				records = append(records, rec)
			}

		case <-res.DoneChan:
			for err := range res.ErrorChan {
				errors = append(errors, err)
			}
			for rec := range res.RecordChan {
				records = append(records, rec)
			}

			return
		}
	}
}

// Check simple search results.
func TestEngineSearchBypass(t *testing.T) {
	SetLogLevelString(testLogLevel)
	taskId = 0 // reset to check intermediate file names

	f1 := newFake(1000, 10)
	f1.Host = "host"

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
			records, errors := drainFull(res)

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

			if assert.EqualValues(t, 1, len(f1.searchDone)) {
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["*.txt"], mode:"", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:"", keep-index:"", delim:"", index:true, data:true}`, f1.searchDone[0].String())
			}
		}
	}
}

// check for simple AND
func TestEngineSearchAnd3(t *testing.T) {
	SetLogLevelString(testLogLevel)
	taskId = 0 // reset to check intermediate file names

	f1 := newFake(1000, 10)
	f1.Host = "host-1"

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
			records, errors := drainFull(res)

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

			if assert.EqualValues(t, 3, len(f1.searchDone)) {
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["1.txt"], mode:"", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-2.txt", keep-index:".work/.temp-idx-dec-00000001-2.txt", delim:"", index:false, data:false}`, f1.searchDone[0].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hell", WIDTH="3")), files:[".work/.temp-dat-dec-00000001-2.txt"], mode:"", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-3.txt", keep-index:".work/.temp-idx-dec-00000001-3.txt", delim:"", index:false, data:false}`, f1.searchDone[1].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("he", WIDTH="3")), files:[".work/.temp-dat-dec-00000001-3.txt"], mode:"", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-4.txt", keep-index:".work/.temp-idx-dec-00000001-4.txt", delim:"", index:false, data:false}`, f1.searchDone[2].String())
			}
		}
	}
}

// check for simple OR
func TestEngineSearchOr3(t *testing.T) {
	SetLogLevelString(testLogLevel)
	taskId = 0 // reset to check intermediate file names

	f1 := newFake(1000, 10)
	f1.Host = "host-1"

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
			records, errors := drainFull(res)

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

			if assert.EqualValues(t, 3, len(f1.searchDone)) {
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hello", WIDTH="3")), files:["1.txt"], mode:"", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-2.txt", keep-index:".work/.temp-idx-dec-00000001-2.txt", delim:"", index:false, data:false}`, f1.searchDone[0].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("hell", WIDTH="3")), files:["1.txt"], mode:"", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-3.txt", keep-index:".work/.temp-idx-dec-00000001-3.txt", delim:"", index:false, data:false}`, f1.searchDone[1].String())
				assert.EqualValues(t, `Config{query:(RAW_TEXT CONTAINS EXACT("he", WIDTH="3")), files:["1.txt"], mode:"", width:3, dist:0, cs:true, nodes:0, limit:0, keep-data:".work/.temp-dat-dec-00000001-4.txt", keep-index:".work/.temp-idx-dec-00000001-4.txt", delim:"", index:false, data:false}`, f1.searchDone[2].String())
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
