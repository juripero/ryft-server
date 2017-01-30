package ryftprim

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test default options
func TestOptions(t *testing.T) {
	// check for good case
	check := func(opts map[string]interface{}) {
		if engine, err := NewEngine(opts); assert.NoError(t, err) {
			assert.EqualValues(t, opts, engine.Options())
		}
	}

	// check for bad case
	bad := func(opts map[string]interface{}, expectedError string) {
		if _, err := NewEngine(opts); assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	// get fake options
	fake := func(name string, val interface{}) map[string]interface{} {
		opts := map[string]interface{}{
			"instance-name":           ".name",
			"ryftprim-exec":           "/bin/false",
			"ryftprim-legacy":         false,
			"ryftprim-kill-on-cancel": true,
			"ryftone-mount":           "/tmp",
			"home-dir":                "/",
			"open-poll":               "100ms",
			"read-poll":               "100ms",
			"read-limit":              50,
			"keep-files":              true,
			"minimize-latency":        true,
			"index-host":              "localhost",
		}

		if len(name) != 0 {
			opts[name] = val
		}

		return opts
	}

	// check default options
	if engine, err := NewEngine(nil); assert.NoError(t, err) {
		assert.EqualValues(t, map[string]interface{}{
			"instance-name":           "",
			"ryftprim-exec":           "/usr/bin/ryftprim",
			"ryftprim-legacy":         true,
			"ryftprim-kill-on-cancel": false,
			"ryftone-mount":           "/ryftone",
			"home-dir":                "/",
			"open-poll":               "50ms",
			"read-poll":               "50ms",
			"read-limit":              100,
			"keep-files":              false,
			"minimize-latency":        false,
			"index-host":              "",
		}, engine.Options())

		assert.EqualValues(t, `ryftprim{instance:"", ryftone:"/ryftone", home:"/", ryftprim:"/usr/bin/ryftprim"}`, engine.String())
	}

	check(fake("home-dir", "/"))

	bad(fake("instance-name", false), `failed to parse "instance-name"`)
	bad(fake("ryftprim-exec", false), `failed to parse "ryftprim-exec"`)
	bad(fake("ryftprim-exec", "/usr/bin/missing-file-name"), `failed to locate ryftprim executable`)
	bad(fake("ryftprim-legacy", []byte{}), `failed to parse "ryftprim-legacy"`)
	bad(fake("ryftprim-kill-on-cancel", []byte{}), `failed to parse "ryftprim-kill-on-cancel"`)
	bad(fake("ryftone-mount", false), `failed to parse "ryftone-mount"`)
	bad(fake("ryftone-mount", "/tmp/missing-dir"), `failed to locate mount point`)
	bad(fake("ryftone-mount", "/bin/false"), `mount point is not a directory`)
	bad(fake("home-dir", false), `failed to parse "home-dir"`)
	bad(fake("ryftone-mount", "/home/"), `failed to create working directory`)
	bad(fake("open-poll", false), `failed to parse "open-poll"`)
	bad(fake("read-poll", false), `failed to parse "read-poll"`)
	bad(fake("read-limit", false), `failed to parse "read-limit"`)
	bad(fake("read-limit", 0), `cannot be negative or zero`)
	bad(fake("keep-files", []byte{}), `failed to parse "keep-files"`)
	bad(fake("minimize-latency", []byte{}), `failed to parse "minimize-latency"`)
	bad(fake("index-host", false), `failed to parse "index-host"`)
}