package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// test safe mode
func TestSafeParseMode(t *testing.T) {
	check := func(s string, ignore bool, skip bool, expectedErrors ...string) {
		m, err := SafeParseMode(s)
		if err != nil {
			for _, e := range expectedErrors {
				assert.Contains(t, err.Error(), e)
			}
		} else {
			assert.EqualValues(t, ignore, m.IsIgnore())
			assert.EqualValues(t, skip, m.IsSkipBusy())
			for _, e := range expectedErrors {
				assert.EqualValues(t, e, m.String())
			}
		}
	}

	check("", false, false, "0s")
	check("ignore", true, false, "ignore")
	check("skip", false, true, "skip")
	check("20s", false, false, "20s")
	check("wait-21s", false, false, "21s")
	check("wait-up-to-22s", false, false, "22s")
	check("-", false, false, "bad timeout", "invalid duration")
	check("-10ms", false, false, "bad timeout", "cannot be negative")
	check("10h", false, false, "bad timeout", "cannot be greater than")
}

// test safe items
func TestSafeItems(t *testing.T) {
	assert.Error(t, SafeSetLogLevelString("bad"))
	assert.NoError(t, SafeSetLogLevelString("debug"))
	assert.Equal(t, "debug", SafeGetLogLevel().String())
	assert.Empty(t, globalSafeItems.items)

	// simple "read" locks
	assert.True(t, SafeLockRead("b", 0))
	assert.False(t, SafeLockWrite("b", 0))
	SafeUnlockRead("b")
	assert.Empty(t, globalSafeItems.items)

	// nested "read" locks
	assert.True(t, SafeLockRead("c", 0))
	assert.True(t, SafeLockRead("c", 0))
	assert.False(t, SafeLockWrite("c", 0))
	SafeUnlockRead("c")
	assert.False(t, SafeLockWrite("c", 0))
	SafeUnlockRead("c")
	assert.Empty(t, globalSafeItems.items)

	// simple "write" locks
	assert.True(t, SafeLockWrite("b", 0))
	assert.False(t, SafeLockRead("b", 0))
	SafeUnlockWrite("b")
	assert.Empty(t, globalSafeItems.items)

	// nested "write" locks
	assert.True(t, SafeLockWrite("c", 0))
	assert.True(t, SafeLockWrite("c", 0))
	assert.False(t, SafeLockRead("c", 0))
	SafeUnlockWrite("c")
	assert.False(t, SafeLockRead("c", 0))
	SafeUnlockWrite("c")
	assert.Empty(t, globalSafeItems.items)

	// unlock unknown
	SafeUnlockRead("d")
	SafeUnlockWrite("d")
	assert.Empty(t, globalSafeItems.items)
}

// test wait safe items
func TestSafeItemsWait(t *testing.T) {
	assert.NoError(t, SafeSetLogLevelString("debug"))
	assert.Empty(t, globalSafeItems.items)

	// simple "read" locks, wait OK
	assert.True(t, SafeLockRead("b", 0))
	go func() {
		time.Sleep(100 * time.Millisecond)
		SafeUnlockRead("b")
	}()
	assert.False(t, SafeLockWrite("b", 0))
	if assert.True(t, SafeLockWrite("b", ShareMode(200*time.Millisecond))) {
		SafeUnlockWrite("b")
	}
	assert.Empty(t, globalSafeItems.items)

	// simple "read" locks, wait FAILED
	assert.True(t, SafeLockRead("b2", 0))
	go func() {
		time.Sleep(300 * time.Millisecond)
		SafeUnlockRead("b2")
	}()
	assert.False(t, SafeLockWrite("b2", 0))
	assert.False(t, SafeLockWrite("b2", ShareMode(200*time.Millisecond)))
	time.Sleep(200 * time.Millisecond) // wait "read" is released
	assert.Empty(t, globalSafeItems.items)

	// nested "read" locks
	assert.True(t, SafeLockRead("c", 0))
	assert.True(t, SafeLockRead("c", ShareMode(100*time.Millisecond)))
	go func() {
		time.Sleep(200 * time.Millisecond)
		SafeUnlockRead("c")
		time.Sleep(200 * time.Millisecond)
		SafeUnlockRead("c")
	}()
	assert.False(t, SafeLockWrite("c", 0))
	if assert.True(t, SafeLockWrite("c", ShareMode(600*time.Millisecond))) {
		SafeUnlockWrite("c")
	}
	assert.Empty(t, globalSafeItems.items)

	// simple "write" locks, wait OK
	assert.True(t, SafeLockWrite("b", 0))
	go func() {
		time.Sleep(100 * time.Millisecond)
		SafeUnlockWrite("b")
	}()
	assert.False(t, SafeLockRead("b", 0))
	if assert.True(t, SafeLockRead("b", ShareMode(200*time.Millisecond))) {
		SafeUnlockRead("b")
	}
	assert.Empty(t, globalSafeItems.items)

	// simple "write" locks, wait FAILED
	assert.True(t, SafeLockWrite("b2", 0))
	go func() {
		time.Sleep(300 * time.Millisecond)
		SafeUnlockWrite("b2")
	}()
	assert.False(t, SafeLockRead("b2", 0))
	assert.False(t, SafeLockRead("b2", ShareMode(200*time.Millisecond)))
	time.Sleep(200 * time.Millisecond) // wait "write" is released
	assert.Empty(t, globalSafeItems.items)

	// nested "write" locks
	assert.True(t, SafeLockWrite("c", 0))
	assert.True(t, SafeLockWrite("c", ShareMode(100*time.Millisecond)))
	go func() {
		time.Sleep(200 * time.Millisecond)
		SafeUnlockWrite("c")
		time.Sleep(200 * time.Millisecond)
		SafeUnlockWrite("c")
	}()
	assert.False(t, SafeLockRead("c", 0))
	if assert.True(t, SafeLockRead("c", ShareMode(600*time.Millisecond))) {
		SafeUnlockRead("c")
	}
	assert.Empty(t, globalSafeItems.items)
}
