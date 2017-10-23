package ryftprim

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

// test Tweaks
func TestTweakOpts(t *testing.T) {
	// parse tweaks from YAML data
	parseYamlTweaks := func(data string) (*Tweaks, error) {
		var cfg map[string]interface{}
		err := yaml.Unmarshal([]byte(data), &cfg)
		if err != nil {
			return nil, err
		}

		return ParseTweaks(cfg)
	}

	// empty config
	opts, err := ParseTweaks(map[string]interface{}{})
	if assert.NoError(t, err) {
		assert.EqualValues(t, []string(nil), opts.GetOptions("default", "ryftx", "es"))
		assert.EqualValues(t, []string(nil), opts.GetOptions("", "ryftprim", "es"))
	}

	check := func(opts []string, expected ...string) {
		assert.EqualValues(t, expected, opts)
	}

	opts, err = parseYamlTweaks(`
backend-tweaks:
  options:
    high.ryftprim.es: [high, prim, es]
    high.ryftprim.ds: [high, prim, ds]
    high.ryftprim: [high, prim]
    high.fhs: [high, fhs]
    high: [high]

    ryftx.es: [x, es]
    ryftx.ts: [x, ts]
    ryftx: [x]
    fhs: [fhs]

    default: ["?"]
`)
	if assert.NoError(t, err) {
		check(opts.GetOptions("high", "ryftprim", "es"), "high", "prim", "es")
		check(opts.GetOptions("high", "ryftprim", "feds"), "high", "prim")
		check(opts.GetOptions("high", "ryftprim", "fhs"), "high", "fhs")
		check(opts.GetOptions("high", "ryftx", "es"), "high")
		check(opts.GetOptions("high", "ryftx", "feds"), "high")
		check(opts.GetOptions("high", "ryftx", "fhs"), "high", "fhs")

		check(opts.GetOptions("", "ryftx", "es"), "x", "es")
		check(opts.GetOptions("", "ryftx", "feds"), "x")
		check(opts.GetOptions("", "ryftx", "fhs"), "fhs")
		check(opts.GetOptions("", "ryftprim", "es"), "?")
		check(opts.GetOptions("", "ryftprim", "feds"), "?")
		check(opts.GetOptions("", "ryftprim", "fhs"), "fhs")
	}
}
