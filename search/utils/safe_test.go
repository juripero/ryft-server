package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// dump safe items
func TestSafeItems(t *testing.T) {
	assert.Empty(t, safeItems)

	// simple "read" locks
	assert.True(t, SafeLockRead("b"))
	assert.False(t, SafeLockWrite("b"))
	SafeUnlockRead("b")
	assert.Empty(t, safeItems)

	// nested "read" locks
	assert.True(t, SafeLockRead("c"))
	assert.True(t, SafeLockRead("c"))
	assert.False(t, SafeLockWrite("c"))
	SafeUnlockRead("c")
	assert.False(t, SafeLockWrite("c"))
	SafeUnlockRead("c")
	assert.Empty(t, safeItems)

	// simple "write" locks
	assert.True(t, SafeLockWrite("b"))
	assert.False(t, SafeLockRead("b"))
	SafeUnlockWrite("b")
	assert.Empty(t, safeItems)

	// nested "write" locks
	assert.True(t, SafeLockWrite("c"))
	assert.True(t, SafeLockWrite("c"))
	assert.False(t, SafeLockRead("c"))
	SafeUnlockWrite("c")
	assert.False(t, SafeLockRead("c"))
	SafeUnlockWrite("c")
	assert.Empty(t, safeItems)

	// unlock unknown
	SafeUnlockRead("d")
	SafeUnlockWrite("d")
	assert.Empty(t, safeItems)
}
