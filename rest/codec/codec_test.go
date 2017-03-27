package codec

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test supported MIME types
func TestCodecMIME(t *testing.T) {
	types := GetSupportedMimeTypes()
	assert.NotEmpty(t, types)
}

// test encoder/decoder
func TestCodecEncoder(t *testing.T) {
	buf := &bytes.Buffer{}

	e1, err := NewEncoder(buf, "application/json", true)
	assert.NoError(t, err)
	assert.NotNil(t, e1)

	e2, err := NewEncoder(buf, "application/json", false)
	assert.NoError(t, err)
	assert.NotNil(t, e2)

	e3, err := NewEncoder(buf, "application/msgpack", true)
	assert.NoError(t, err)
	assert.NotNil(t, e3)

	e4, err := NewEncoder(buf, "application/msgpack", false)
	assert.NoError(t, err)
	assert.NotNil(t, e4)

	e5, err := NewEncoder(buf, "text/csv", true)
	assert.NoError(t, err)
	assert.NotNil(t, e5)

	e6, err := NewEncoder(buf, "text/csv", false)
	assert.NoError(t, err)
	assert.NotNil(t, e6)

	_, err = NewEncoder(buf, "application/octet-stream", true)
	assert.Error(t, err)

	_, err = NewDecoder(buf, "application/json", true)
	assert.Error(t, err)
}
