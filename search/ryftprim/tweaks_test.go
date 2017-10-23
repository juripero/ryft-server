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
		assert.EqualValues(t, []string{"high", "prim", "es"}, opts.GetOptions("high", "ryftprim", "es"))
		assert.EqualValues(t, []string{"high", "prim"}, opts.GetOptions("high", "ryftprim", "feds"))
		assert.EqualValues(t, []string{"high", "fhs"}, opts.GetOptions("high", "ryftprim", "fhs"))
		assert.EqualValues(t, []string{"high"}, opts.GetOptions("high", "ryftx", "es"))
		assert.EqualValues(t, []string{"high"}, opts.GetOptions("high", "ryftx", "feds"))
		assert.EqualValues(t, []string{"high", "fhs"}, opts.GetOptions("high", "ryftx", "fhs"))

		assert.EqualValues(t, []string{"x", "es"}, opts.GetOptions("", "ryftx", "es"))
		assert.EqualValues(t, []string{"x"}, opts.GetOptions("", "ryftx", "feds"))
		assert.EqualValues(t, []string{"fhs"}, opts.GetOptions("", "ryftx", "fhs"))
		assert.EqualValues(t, []string{"?"}, opts.GetOptions("", "ryftprim", "es"))
		assert.EqualValues(t, []string{"?"}, opts.GetOptions("", "ryftprim", "feds"))
		assert.EqualValues(t, []string{"fhs"}, opts.GetOptions("", "ryftprim", "fhs"))
	}

	opts, err = parseYamlTweaks(`
backend-tweaks:
  options:
    default: ["?"]
  router:
    pcre2: ryftpcre2
    fhs,feds: ryftprim
    default: ryftx
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, "ryftpcre2", opts.GetBackendTool("pcre2"))
		assert.EqualValues(t, "ryftprim", opts.GetBackendTool("feds"))
		assert.EqualValues(t, "ryftprim", opts.GetBackendTool("fhs"))
		assert.EqualValues(t, "ryftx", opts.GetBackendTool("es"))
		assert.EqualValues(t, "ryftx", opts.GetBackendTool("ts"))
	}
}
