package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test empty configuration
func TestConfigEmpty(t *testing.T) {
	cfg := NewEmptyConfig()
	assert.Empty(t, cfg.Query, "no query expected")
	assert.Empty(t, cfg.Files, "no files expected")
	assert.Empty(t, cfg.Mode)

	assert.Empty(t, cfg.KeepDataAs)
	assert.Empty(t, cfg.KeepIndexAs)
	assert.Empty(t, cfg.Delimiter)

	assert.Equal(t, `Config{query:, files:[], mode:"", width:0, dist:0, cs:false, nodes:0, limit:0, keep-data:"", keep-index:"", delim:"", index:false, data:false}`, cfg.String())
}

// test simple configuration
func TestConfigSimple(t *testing.T) {
	cfg := NewConfig("hello", "a.txt", "b.txt")
	assert.Equal(t, "hello", cfg.Query)
	assert.Equal(t, []string{"a.txt", "b.txt"}, cfg.Files)
	assert.Empty(t, cfg.Mode)

	cfg.AddFile("c.txt", "d.txt")
	assert.Equal(t, []string{"a.txt", "b.txt", "c.txt", "d.txt"}, cfg.Files)

	assert.Empty(t, cfg.KeepDataAs)
	assert.Empty(t, cfg.KeepIndexAs)
	assert.Empty(t, cfg.Delimiter)

	cfg.Mode = "fhs"
	cfg.Delimiter = "\r\n\f"
	cfg.ReportIndex = true
	cfg.ReportData = true
	assert.Equal(t, `Config{query:hello, files:["a.txt" "b.txt" "c.txt" "d.txt"], mode:"fhs", width:0, dist:0, cs:false, nodes:0, limit:0, keep-data:"", keep-index:"", delim:"\r\n\f", index:true, data:true}`, cfg.String())
}
