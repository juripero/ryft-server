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
	assert.Empty(t, cfg.KeepViewAs)
	assert.Empty(t, cfg.Delimiter)

	assert.Equal(t, `Config{cs:true}`, cfg.String())
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
	assert.Equal(t, `Config{query:hello, files:["a.txt" "b.txt" "c.txt" "d.txt"], mode:"fhs", cs:true, delim:#0d0a0c, I, D}`, cfg.String())
}

// test relative to home
func TestConfigRelativeToHome(t *testing.T) {
	cfg := NewConfig("hello", "../a.txt", "../b.txt")
	cfg.KeepIndexAs = "../index.txt"
	cfg.KeepDataAs = "../data.txt"
	cfg.KeepViewAs = "../view.txt"

	// input
	if err := cfg.CheckRelativeToHome("/ryftone"); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "path")
		assert.Contains(t, err.Error(), "is not relative to home")
	}
	cfg.Files = nil

	// index
	if err := cfg.CheckRelativeToHome("/ryftone"); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "index")
		assert.Contains(t, err.Error(), "is not relative to home")
	}
	cfg.KeepIndexAs = ""

	// data
	if err := cfg.CheckRelativeToHome("/ryftone"); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "data")
		assert.Contains(t, err.Error(), "is not relative to home")
	}
	cfg.KeepDataAs = ""

	// view
	if err := cfg.CheckRelativeToHome("/ryftone"); assert.Error(t, err) {
		assert.Contains(t, err.Error(), "view")
		assert.Contains(t, err.Error(), "is not relative to home")
	}
	cfg.KeepViewAs = ""

	// valid
	assert.NoError(t, cfg.CheckRelativeToHome("/ryftone"))
}
