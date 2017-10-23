package ryftprim

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/stretchr/testify/assert"
)

var (
	testLogLevel = "error"
)

// should be formatted with root directory
const testFakeRyftprimScript3 = `#!/bin/bash
# test script to emulate ryftprim

# initial delay
sleep 1s

OD=%[1]s/ryftprim/ryftprim-data.bin
OI=%[1]s/ryftprim/ryftprim-index.txt
DELIM=$'\r\n\f'
mkdir -p %[1]s/ryftprim

# parse options
while [[ $# > 0 ]]; do
	case "$1" in
	-od)
		OD="%[1]s/$2"
		shift 2
		;;
	-oi)
		OI="%[1]s/$2"
		shift 2
		;;
	*) # unknown option, skip it
		shift 1
		;;
	esac
done

# first record
echo -n "hello" > "$OD"
echo -n "$DELIM" >> "$OD"
echo "%[1]s/ryftprim/1.txt,100,5,0" > "$OI"

# second record
echo "2.txt,200,5,n/a" >> "$OI" # FALLBACK to absolute
sleep 0.1s                      # emulate "no data"
echo -n "hello" >> "$OD"
echo -n "$DELIM" >> "$OD"

# third record
echo -n "hello" >> "$OD"
echo -n "$DELIM" >> "$OD"
sleep 0.1s
echo -n "%[1]s/ryftprim/" >> "$OI"   # first INDEX part
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
`

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

// write a script (executable file)
func testWriteScript3(path string, root string) error {
	return testWriteScript(path, fmt.Sprintf(testFakeRyftprimScript3, root))
}

// get files
func TestEngineFiles(t *testing.T) {
	SetLogLevelString(testLogLevel)

	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	assert.NoError(t, os.MkdirAll(filepath.Join(root, "test/dir"), 0755))
	assert.NoError(t, testWriteScript(filepath.Join(root, "test/1.txt"), "1111"))
	defer os.RemoveAll(root)

	engine, err := factory(map[string]interface{}{
		"instance-name":           ".test",
		"ryftprim-exec":           "/bin/false",
		"ryftprim-legacy":         true,
		"ryftprim-kill-on-cancel": true,
		"ryftone-mount":           root,
		"home-dir":                "/test/",
		"minimize-latency":        true,
		"index-host":              "hozt",
	})
	if !assert.NoError(t, err) {
		return
	}

	res, err := engine.Files("/", false)
	if !assert.NoError(t, err) {
		return
	}

	assert.NotEmpty(t, res.Dirs)
	assert.NotEmpty(t, res.Files)

	// fail on missing directory
	_, err = engine.Files("missing-tmp-dir", false)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "failed to read directory content")
	}

	// fail on bad file
	_, err = engine.Files("../dir", false)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "is not relative to user's home")
	}
}

// test log level
func TestLogLevel(t *testing.T) {
	if err := SetLogLevelString("bad-log-level"); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "not a valid logrus Level")
	}
	SetLogLevelString(testLogLevel)
	assert.EqualValues(t, testLogLevel, GetLogLevel().String())
}

// valid
func TestEngineUsual(t *testing.T) {
	SetLogLevelString(testLogLevel)

	// prepare ryftprim emulation script
	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	prim := fmt.Sprintf("/tmp/ryftprim-%x.sh", time.Now().UnixNano())
	assert.NoError(t, testWriteScript3(prim, root))
	assert.NoError(t, os.MkdirAll(root, 0755))
	defer os.RemoveAll(prim)
	defer os.RemoveAll(root)

	check := func(mode string) {
		cfg := search.NewConfig("hello", "1.txt", "2.txt")
		cfg.Mode = mode
		cfg.Width = 10
		cfg.Case = false
		cfg.Dist = 1
		cfg.KeepIndexAs = "ryftprim-index" // NO ".txt" extension
		cfg.KeepDataAs = "ryftprim-data.bin"
		cfg.Delimiter = "\r\n\f"
		cfg.ReportIndex = true
		cfg.ReportData = true
		cfg.Nodes = 1

		engine, err := factory(map[string]interface{}{
			"instance-name":           ".test",
			"ryftprim-exec":           prim,
			"ryftprim-legacy":         true,
			"ryftprim-kill-on-cancel": true,
			"ryftone-mount":           root,
			"home-dir":                "ryftprim",
			"minimize-latency":        true,
			"index-host":              "hozt",
			"backend-tweaks": map[string]interface{}{
				"options": map[string][]string{},
				"router":  map[string]string{},
			},
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

	check("") // generic
	check("es")
	os.RemoveAll(filepath.Join(root, "ryfptrim/ryftprim-data.bin")) // postpone processing
	check("fhs")
	os.RemoveAll(filepath.Join(root, "ryfptrim/ryftprim-index.txt")) // postpone processing
	check("feds")
	check("ds")
	check("ts")
	check("ns")
	check("cs")
	check("ipv4")
	check("ipv6")
}

// valid results (limit to 2)
func TestEngineUsualLimit(t *testing.T) {
	SetLogLevelString(testLogLevel)

	// prepare ryftprim emulation script
	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	prim := fmt.Sprintf("/tmp/ryftprim-%x.sh", time.Now().UnixNano())
	assert.NoError(t, testWriteScript3(prim, root))
	assert.NoError(t, os.MkdirAll(root, 0755))
	defer os.RemoveAll(prim)
	defer os.RemoveAll(root)

	cfg := search.NewConfig("hello", "1.txt", "2.txt")
	cfg.KeepIndexAs = "ryftprim-index.txt"
	cfg.KeepDataAs = "ryftprim-data.bin"
	cfg.Delimiter = "\r\n\f"
	cfg.ReportIndex = true
	cfg.ReportData = true
	cfg.Nodes = 1
	cfg.Limit = 2

	engine, err := factory(map[string]interface{}{
		"instance-name":           ".test",
		"ryftprim-exec":           prim,
		"ryftprim-legacy":         true,
		"ryftprim-kill-on-cancel": true,
		"ryftone-mount":           root,
		"home-dir":                "ryftprim",
		"minimize-latency":        true,
		"index-host":              "hozt",
		"backend-tweaks": map[string]interface{}{
			"options": map[string][]string{},
			"router":  map[string]string{},
		},
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
		assert.EqualValues(t, 2, res.RecordsReported()) {

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
	}
}

// valid (no output files)
func TestEngineUsualNoOutput(t *testing.T) {
	SetLogLevelString(testLogLevel)

	// prepare ryftprim emulation script
	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	prim := fmt.Sprintf("/tmp/ryftprim-%x.sh", time.Now().UnixNano())
	assert.NoError(t, testWriteScript3(prim, root))
	assert.NoError(t, os.MkdirAll(root, 0755))
	defer os.RemoveAll(prim)
	defer os.RemoveAll(root)

	cfg := search.NewConfig("hello", "1.txt", "2.txt")
	cfg.Mode = "fhs"
	cfg.Width = -1
	cfg.Case = false
	cfg.Reduce = true
	cfg.Dist = 1
	//cfg.KeepIndexAs = "ryftprim-index.txt"
	//cfg.KeepDataAs = "ryftprim-data.bin"
	cfg.Delimiter = "\r\n\f"
	cfg.ReportIndex = true
	cfg.ReportData = true
	cfg.Nodes = 1

	engine, err := factory(map[string]interface{}{
		"instance-name":           ".test",
		"ryftprim-exec":           prim,
		"ryftprim-legacy":         true,
		"ryftprim-kill-on-cancel": true,
		"ryftone-mount":           root,
		"home-dir":                "ryftprim",
		"minimize-latency":        true,
		"index-host":              "hozt",
		"backend-tweaks": map[string]interface{}{
			"options": map[string][]string{},
			"router":  map[string]string{},
		},
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

// bad search mode
func TestEngineBadSearchMode(t *testing.T) {
	SetLogLevelString(testLogLevel)

	cfg := search.NewConfig("hello", "1.txt", "2.txt")
	cfg.Mode = "bad"

	engine, err := factory(map[string]interface{}{
		"instance-name":           ".test",
		"ryftprim-exec":           "/bin/false",
		"ryftprim-legacy":         true,
		"ryftprim-kill-on-cancel": true,
		"ryftone-mount":           "/tmp",
		"home-dir":                "ryft",
		"minimize-latency":        true,
		"index-host":              "hozt",
		"backend-tweaks": map[string]interface{}{
			"options": map[string][]string{},
			"router":  map[string]string{},
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	_, err = engine.Search(cfg)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "is unknown search mode")
	}

	cfg.Mode = "fhs"
	cfg.ReportData = true
	cfg.ReportIndex = false
	_, err = engine.Search(cfg)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "failed to report DATA without INDEX")
	}
}

// failed to start (bad path)
func TestEngineBadPath(t *testing.T) {
	SetLogLevelString(testLogLevel)

	// prepare ryftprim emulation script
	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	prim := fmt.Sprintf("/tmp/ryftprim-%x.sh", time.Now().UnixNano())
	assert.NoError(t, testWriteScript3(prim, root))
	assert.NoError(t, os.MkdirAll(root, 0755))
	defer os.RemoveAll(prim)
	defer os.RemoveAll(root)

	cfg := search.NewConfig("hello")
	cfg.Mode = "fhs"

	engine, err := factory(map[string]interface{}{
		"instance-name":           ".test",
		"ryftprim-exec":           prim,
		"ryftprim-legacy":         true,
		"ryftprim-kill-on-cancel": true,
		"ryftone-mount":           root,
		"home-dir":                "ryftprim",
		"minimize-latency":        true,
		"index-host":              "hozt",
		"backend-tweaks": map[string]interface{}{
			"options": map[string][]string{},
			"router":  map[string]string{},
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	// bad input
	cfg.Files = []string{"../1.txt", "../2.txt"}
	_, err = engine.Search(cfg)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "is not relative to home")
	}
	cfg.Files = []string{"1.txt", "2.txt"}

	// bad index
	cfg.KeepIndexAs = "../../etc/index.txt"
	_, err = engine.Search(cfg)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "is not relative to home")
	}
	cfg.KeepIndexAs = ""

	// bad data
	cfg.KeepDataAs = "../data.txt"
	_, err = engine.Search(cfg)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "is not relative to home")
	}
	cfg.KeepDataAs = ""
}

// failed to start tool
func TestEngineFailedToStartTool(t *testing.T) {
	SetLogLevelString(testLogLevel)

	// prepare ryftprim emulation script
	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	prim := fmt.Sprintf("/tmp/ryftprim-%x.sh", time.Now().UnixNano())
	assert.NoError(t, testWriteScript3(prim, root))
	assert.NoError(t, os.MkdirAll(root, 0755))
	// defer os.RemoveAll(prim)
	defer os.RemoveAll(root)

	cfg := search.NewConfig("hello", "1.txt", "2.txt")
	cfg.Mode = "fhs"

	engine, err := factory(map[string]interface{}{
		"instance-name":           ".test",
		"ryftprim-exec":           prim,
		"ryftprim-legacy":         true,
		"ryftprim-kill-on-cancel": true,
		"ryftone-mount":           root,
		"home-dir":                "ryftprim",
		"minimize-latency":        true,
		"index-host":              "hozt",
		"backend-tweaks": map[string]interface{}{
			"options": map[string][]string{},
			"router":  map[string]string{},
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	os.RemoveAll(prim)

	_, err = engine.Search(cfg)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "failed to start tool")
	}
}

// cancel tool
func TestEngineCancelTool(t *testing.T) {
	SetLogLevelString(testLogLevel)

	// prepare ryftprim emulation script
	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	prim := fmt.Sprintf("/tmp/ryftprim-%x.sh", time.Now().UnixNano())
	if err := testWriteScript(prim,
		`#!/bin/bash
sleep 300s
`); !assert.NoError(t, err) {
		return
	}
	assert.NoError(t, os.MkdirAll(root, 0755))

	defer os.RemoveAll(prim)
	defer os.RemoveAll(root)

	cfg := search.NewConfig("hello", "1.txt", "2.txt")
	cfg.ReportIndex = true

	engine, err := factory(map[string]interface{}{
		"instance-name":           ".test",
		"ryftprim-exec":           prim,
		"ryftprim-legacy":         true,
		"ryftprim-kill-on-cancel": true,
		"ryftone-mount":           root,
		"home-dir":                "ryftprim",
		"minimize-latency":        true,
		"index-host":              "hozt",
		"backend-tweaks": map[string]interface{}{
			"options": map[string][]string{},
			"router":  map[string]string{},
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	res, err := engine.Search(cfg)
	if !assert.NoError(t, err) {
		return
	}

	time.Sleep(200 * time.Millisecond)
	res.JustCancel() // cancel by user

	<-res.DoneChan // wait results

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 0, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {
		assert.True(t, res.IsCancelled())
	}
}

// tool failed
func TestEngineToolFailed(t *testing.T) {
	SetLogLevelString(testLogLevel)

	cfg := search.NewConfig("hello", "1.txt", "2.txt")
	cfg.ReportIndex = true

	engine, err := factory(map[string]interface{}{
		"instance-name":           ".test",
		"ryftprim-exec":           "/bin/false",
		"ryftprim-legacy":         false,
		"ryftprim-kill-on-cancel": true,
		"ryftone-mount":           "/tmp",
		"home-dir":                "ryft",
		"minimize-latency":        true,
		"index-host":              "hozt",
		"backend-tweaks": map[string]interface{}{
			"options": map[string][]string{},
			"router":  map[string]string{},
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	res, err := engine.Search(cfg)
	if !assert.NoError(t, err) {
		return
	}

	time.Sleep(200 * time.Millisecond)
	res.JustCancel() // cancel by user

	<-res.DoneChan // wait results

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 1, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {
		assert.True(t, res.IsCancelled())

		// check first error
		if err := <-res.ErrorChan; assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "ryftprim failed with exit status")
		}
	}
}

// tool failed (empty input data set)
func TestEngineToolFailed2(t *testing.T) {
	SetLogLevelString(testLogLevel)

	// prepare ryftprim emulation script
	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	prim := fmt.Sprintf("/tmp/ryftprim-%x.sh", time.Now().UnixNano())
	if err := testWriteScript(prim,
		`#!/bin/bash
echo "bla bla bla ERROR:  Input data set cannot be empty bla bla bla"
exit(3)
`); !assert.NoError(t, err) {
		return
	}
	assert.NoError(t, os.MkdirAll(root, 0755))
	defer os.RemoveAll(prim)
	defer os.RemoveAll(root)

	cfg := search.NewConfig("hello", "1.txt", "2.txt")
	cfg.ReportIndex = true

	engine, err := factory(map[string]interface{}{
		"instance-name":           ".test",
		"ryftprim-exec":           prim,
		"ryftprim-legacy":         true,
		"ryftprim-kill-on-cancel": true,
		"ryftone-mount":           root,
		"home-dir":                "ryftprim",
		"minimize-latency":        true,
		"index-host":              "hozt",
		"backend-tweaks": map[string]interface{}{
			"options": map[string][]string{},
			"router":  map[string]string{},
		},
	})
	if !assert.NoError(t, err) {
		return
	}

	res, err := engine.Search(cfg)
	if !assert.NoError(t, err) {
		return
	}

	time.Sleep(200 * time.Millisecond)
	res.JustCancel() // cancel by user

	<-res.DoneChan // wait results

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 1, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {
		assert.True(t, res.IsCancelled())

		// check first error
		if err := <-res.ErrorChan; assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "ryftprim failed with exit status")
		}
	}
}

func TestTweakOpts(t *testing.T) {
	// empty config
	data := map[string][]string{}
	opts := NewTweakOpts(data)
	assert.Equal(t, []string{}, opts.GetOptions("normal", "ryftx", "es"))

	// longest chain wins
	data = map[string][]string{
		"normal.ryftx.es": []string{"1"},
		"normal.ryftx":    []string{"2"},
		"normal":          []string{"3"},
	}
	opts = NewTweakOpts(data)
	assert.Equal(t, data["normal.ryftx.es"], opts.GetOptions("normal", "ryftx", "es"))
	assert.Equal(t, data["normal.ryftx"], opts.GetOptions("normal", "ryftx", "time"))
	assert.Equal(t, data["normal.ryftx"], opts.GetOptions("normal", "ryftx", ""))
	assert.Equal(t, data["normal"], opts.GetOptions("normal", "ryftprim", "es"))
	assert.Equal(t, data["normal"], opts.GetOptions("normal", "", ""))

	// backend wins a mode
	data = map[string][]string{
		"normal.ryftx.es": []string{"1"},
		"normal.ryftx":    []string{"2"},
		"normal":          []string{"3"},
		"ryftx":           []string{"4"},
	}
	opts = NewTweakOpts(data)

	assert.Equal(t, data["normal.ryftx.es"], opts.GetOptions("normal", "ryftx", "es"))
	assert.Equal(t, data["normal.ryftx"], opts.GetOptions("normal", "ryftx", ""))
	assert.Equal(t, data["normal"], opts.GetOptions("normal", "ryftprim", ""))
	assert.Equal(t, data["ryftx"], opts.GetOptions("", "ryftx", "ds"))
	assert.Equal(t, data["ryftx"], opts.GetOptions("", "ryftx", "es"))

	// primitive wins backend
	data = map[string][]string{
		"normal.ryftx.es": []string{"1"},
		"normal.ryftx":    []string{"2"},
		"ryftx":           []string{"4"},
		"ryftx.es":        []string{"5"},
	}
	opts = NewTweakOpts(data)
	assert.Equal(t, data["ryftx"], opts.GetOptions("", "ryftx", "ds"))
	assert.Equal(t, data["ryftx.es"], opts.GetOptions("", "ryftx", "es"))

	// choose mode
	data = map[string][]string{
		"normal":       []string{"1"},
		"hp":           []string{"2"},
		"normal.ryftx": []string{"3"},
		"ryftx":        []string{"4"},
	}
	opts = NewTweakOpts(data)
	assert.Equal(t, data["normal"], opts.GetOptions("normal", "ryftprim", ""))
	assert.Equal(t, data["normal.ryftx"], opts.GetOptions("normal", "ryftx", ""))
	assert.Equal(t, data["ryftx"], opts.GetOptions("", "ryftx", ""))
	assert.Equal(t, data["hp"], opts.GetOptions("hp", "ryftprim", ""))
	assert.Equal(t, []string{}, opts.GetOptions("", "ryftprim", ""))

	// primitive wins all
	data = map[string][]string{
		"normal.ryftx":    []string{"1"},
		"normal":          []string{"2"},
		"ryftx":           []string{"3"},
		"normal.ryftx.es": []string{"4"},
		"es":              []string{"5"},
	}
	opts = NewTweakOpts(data)
	assert.Equal(t, data["es"], opts.GetOptions("", "", "es"))
	assert.Equal(t, data["es"], opts.GetOptions("normal", "ryftprim", "es"))
	assert.Equal(t, data["normal.ryftx.es"], opts.GetOptions("normal", "ryftx", "es"))
}
