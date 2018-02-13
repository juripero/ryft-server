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

	// abs-path (as a map)
	opts, err := parseYamlTweaks(`
backend-tweaks:
  abs-path:
    ryftprim: false
    ryftx: true
    default: false
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, map[string]bool{
			"ryftprim": false,
			"ryftx":    true,
			"default":  false,
		}, opts.UseAbsPath)
	}

	// abs-path (as a slice)
	opts, err = parseYamlTweaks(`
backend-tweaks:
  abs-path:
  - ryftx
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, map[string]bool{
			"ryftx": true,
		}, opts.UseAbsPath)
	}
	opts, err = parseYamlTweaks(`
backend-tweaks:
  abs-path: [ ryftx ]
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, map[string]bool{
			"ryftx": true,
		}, opts.UseAbsPath)
	}
	opts, err = parseYamlTweaks(`
backend-tweaks:
  abs-path: [ ryftx, ryftpcre2 ]
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, map[string]bool{
			"ryftx":     true,
			"ryftpcre2": true,
		}, opts.UseAbsPath)
	}

	// abs-path (as a string)
	opts, err = parseYamlTweaks(`
backend-tweaks:
  abs-path: ryftx
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, map[string]bool{
			"ryftx": true,
		}, opts.UseAbsPath)
	}

	// abs-path (as a bool)
	opts, err = parseYamlTweaks(`
backend-tweaks:
  abs-path: true
`)
	if assert.NoError(t, err) {
		assert.EqualValues(t, map[string]bool{
			"default": true,
		}, opts.UseAbsPath)
	}

	// abs-path fails
	opts, err = parseYamlTweaks(`
backend-tweaks:
  abs-path: 100
`)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), `unknown "backend-tweaks.abs-path" option type`)
	}
}
