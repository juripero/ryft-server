package ryfthttp

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
			"server-url": "http://localhost:8765",
			"auth-token": "Basic: Login+Password",
			"local-only": true,
			"skip-stat":  false,
			"index-host": "localhost",
		}

		if len(name) != 0 {
			opts[name] = val
		}

		return opts
	}

	// check default options
	if engine, err := NewEngine(nil); assert.NoError(t, err) {
		assert.EqualValues(t, map[string]interface{}{
			"server-url": "http://localhost:8765",
			"auth-token": "",
			"local-only": false,
			"skip-stat":  false,
			"index-host": "",
		}, engine.Options())

		assert.EqualValues(t, `ryfthttp{url:"http://localhost:8765", local:false, stat:true}`, engine.String())
	}

	check(fake("server-url", "http://localhost:8765"))

	bad(fake("server-url", false), `failed to parse "server-url"`)
	bad(fake("server-url", ":bad"), `failed to parse "server-url"`)
	bad(fake("auth-token", false), `auth-token"`)
	bad(fake("local-only", []byte{}), `failed to parse "local-only"`)
	bad(fake("skip-stat", []byte{}), `failed to parse "skip-stat"`)
	bad(fake("index-host", false), `failed to parse "index-host"`)
}
