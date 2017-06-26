package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// convert to JSON
func toJson(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(data)
}

// test user config override
func TestUserConfig(t *testing.T) {
	var server Server

	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	os.MkdirAll(filepath.Join(root, "test"), 0755)
	defer os.RemoveAll(root)

	server.Config.DefaultUserConfig = map[string]interface{}{
		"record-queries": map[string]interface{}{
			"enabled": false,
			"xml":     []string{"*.xml"},
			"csv":     []string{"*.csv"},
		},
	}
	server.Config.BackendOptions = map[string]interface{}{
		"ryftone-mount": root,
	}

	cfg, err := server.getUserConfig("test")
	if assert.NoError(t, err) {
		assert.JSONEq(t, `{"record-queries": {"enabled":false, "xml":["*.xml"], "csv":["*.csv"]}}`, toJson(cfg))
	}

	ioutil.WriteFile(filepath.Join(root, "test/.ryft-user.json"),
		[]byte(`
{"record-queries": {
	"enabled":true,
	"xml":["*.xml1"],
	"csv":["*.csv1"]
}}`), 0644)
	cfg, err = server.getUserConfig("test")
	if assert.NoError(t, err) {
		assert.JSONEq(t, `{"record-queries": {"enabled":true, "xml":["*.xml1"], "csv":["*.csv1"]}}`, toJson(cfg))
	}

	ioutil.WriteFile(filepath.Join(root, "test/.ryft-user.yaml"),
		[]byte(`# YAML config
record-queries:
  enabled: false
  xml: ["*.xml2"]
  csv: ["*.csv2"]
`), 0644)
	cfg, err = server.getUserConfig("test")
	if assert.NoError(t, err) {
		assert.JSONEq(t, `{"record-queries": {"enabled":false, "xml":["*.xml2"], "csv":["*.csv2"]}}`, toJson(cfg))
	}

}
