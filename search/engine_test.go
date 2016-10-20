package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test bad engine
func TestEngineBad(t *testing.T) {
	_, err := NewEngine("bad-engine-name", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is unknown search engine")
}

// test register factory
func TestEngineRegister(t *testing.T) {
	name := "test-engine"

	// register fake engine
	RegisterEngine(name, func(map[string]interface{}) (Engine, error) { return nil, nil })
	assert.Equal(t, []string{name}, GetAvailableEngines())

	// create engine by name
	engine, err := NewEngine(name, nil)
	assert.NoError(t, err)
	assert.Nil(t, engine) // because factory returns nil

	// unregister fake engine
	RegisterEngine(name, nil)
	assert.Empty(t, GetAvailableEngines())
}
