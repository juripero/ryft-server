package rest

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetenvFallback(t *testing.T) {
	result := GetenvFallback("test_get_env_fallback_variable", "got it!")
	assert.Equal(t, "got it!", result)
	os.Setenv("test_get_env_fallback_variable", "any value")

	result = GetenvFallback("test_get_env_fallback_variable", "got it!")
	assert.Equal(t, "any value", result)
}
