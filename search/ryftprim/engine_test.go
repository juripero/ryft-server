package ryftprim

import (
	"os"
	"testing"

	"github.com/getryft/ryft-server/search"
	"github.com/stretchr/testify/assert"
)

// write a script (executable file)
func testWriteScript(path string, script string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(script)
	return nil // OK
}

// valid results
func TestEngineUsual(t *testing.T) {
	SetLogLevelString(testLogLevel)

	// prepare ryftprim emulation script
	if err := testWriteScript("/tmp/ryftprim.sh",
		`#!/bin/bash
# test script to emulate ryftprim

# initial delay
sleep 1s

OD=/tmp/ryft/ryftprim-data.bin
OI=/tmp/ryft/ryftprim-index.txt
DELIM=$'\r\n\f'
mkdir -p /tmp/ryft

# first record
echo -n "hello" > "$OD"
echo -n "$DELIM" >> "$OD"
echo "/tmp/ryft/1.txt,100,5,0" > "$OI"

# second record
echo "2.txt,200,5,n/a" >> "$OI" # FALLBACK to absolute
sleep 0.1s                      # emulate "no data"
echo -n "hello" >> "$OD"
echo -n "$DELIM" >> "$OD"

# third record
echo -n "hello" >> "$OD"
echo -n "$DELIM" >> "$OD"
sleep 0.1s
echo -n "/tmp/ryft/" >> "$OI"   # first INDEX part
sleep 0.1s
echo -n "3.txt,300,5" >> "$OI" # second INDEX part
sleep 0.1s
echo ",1" >> "$OI"             # last INDEX part

# print statistics
echo "Matches: 3"
echo "Duration: 100"
echo "Total Bytes: 1024"
echo "Data Rate: 100 MB/sec"
echo "Fabric Data Rate: 200 MB/sec"
`); !assert.NoError(t, err) {
		return
	}

	defer os.RemoveAll("/tmp/ryftprim.sh")
	defer os.RemoveAll("/tmp/ryft")

	cfg := search.NewConfig("hello", "1.txt", "2.txt")
	cfg.KeepIndexAs = "ryftprim-index.txt"
	cfg.KeepDataAs = "ryftprim-data.bin"
	cfg.Delimiter = "\r\n\f"
	cfg.ReportIndex = true
	cfg.ReportData = true

	engine, err := factory(map[string]interface{}{
		"instance-name":           ".test",
		"ryftprim-exec":           "/tmp/ryftprim.sh",
		"ryftprim-legacy":         true,
		"ryftprim-kill-on-cancel": true,
		"ryftone-mount":           "/tmp",
		"home-dir":                "ryft",
		"minimize-latency":        true,
		"index-host":              "hozt",
	})
	if !assert.NoError(t, err) {
		return
	}

	res, err := engine.Search(cfg)
	if !assert.NoError(t, err) {
		return
	}

	<-res.DoneChan // wait results

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 0, res.ErrorsReported()) &&
		assert.EqualValues(t, 3, res.RecordsReported()) {

		// check first record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "1.txt", rec.Index.File)
			assert.EqualValues(t, 100, rec.Index.Offset)
			assert.EqualValues(t, 5, rec.Index.Length)
			assert.EqualValues(t, 0, rec.Index.Fuzziness)
			assert.EqualValues(t, "hello", rec.RawData)
		}

		// check second record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "2.txt", rec.Index.File)
			assert.EqualValues(t, 200, rec.Index.Offset)
			assert.EqualValues(t, 5, rec.Index.Length)
			assert.EqualValues(t, -1, rec.Index.Fuzziness)
			assert.EqualValues(t, "hello", rec.RawData)
		}

		// check third record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "3.txt", rec.Index.File)
			assert.EqualValues(t, 300, rec.Index.Offset)
			assert.EqualValues(t, 5, rec.Index.Length)
			assert.EqualValues(t, 1, rec.Index.Fuzziness)
			assert.EqualValues(t, "hello", rec.RawData)
		}
	}
}
