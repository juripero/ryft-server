/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2015, Ryft Systems, Inc.
 * All rights reserved.
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *   this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *   this list of conditions and the following disclaimer in the documentation and/or
 *   other materials provided with the distribution.
 * 3. All advertising materials mentioning features or use of this software must display the following acknowledgement:
 *   This product includes software developed by Ryft Systems, Inc.
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used
 *   to endorse or promote products derived from this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY RYFT SYSTEMS, INC. ''AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL RYFT SYSTEMS, INC. BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 * ============
 */

package utils

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

var (
	// logger instance
	safeLog = logrus.New()

	SAFE = "safe"
)

// MaxWaitTimeout maximum allowed timeout
var MaxWaitTimeout = 30 * time.Second

const (
	// we can use negative values for special flags:
	SM_IGNORE ShareMode = -1 // special value for "ignore" flag
	SM_SKIP   ShareMode = -2 // special value for "skip" flag
)

// SafeSetLogLevelString changes global module log level.
func SafeSetLogLevelString(level string) error {
	ll, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}

	SafeSetLogLevel(ll)
	return nil // OK
}

// SafeSetLogLevel changes global module log level.
func SafeSetLogLevel(level logrus.Level) {
	safeLog.Level = level
}

// SafeGetLogLevel gets global module log level.
func SafeGetLogLevel() logrus.Level {
	return safeLog.Level
}

// ShareMode is share mode
type ShareMode time.Duration

// SafeParseMode checks the share mode is OK
func SafeParseMode(mode string) (ShareMode, error) {
	sm := strings.ToLower(strings.TrimSpace(mode))

	switch sm {
	case "ignore", "force-ignore":
		return SM_IGNORE, nil
	case "skip", "skip-busy":
		return SM_SKIP, nil
	case "":
		return 0, nil // don't wait
	}

	sm = strings.TrimPrefix(sm, "wait-up-to-")
	sm = strings.TrimPrefix(sm, "wait-")
	if d, err := time.ParseDuration(sm); err != nil {
		return 0, fmt.Errorf("bad timeout: %s", err)
	} else if d < 0 {
		return 0, fmt.Errorf("bad timeout: cannot be negative, found %s", d)
	} else if d > MaxWaitTimeout {
		return 0, fmt.Errorf("bad timeout: cannet be > %s, found %s", MaxWaitTimeout, d)
	} else {
		return ShareMode(d), nil // OK
	}
}

// IsIgnore checks if share mode is "ignore".
func (sm ShareMode) IsIgnore() bool {
	return sm == SM_IGNORE
}

// IsSkipBusy checks if share mode is "skip-busy".
func (sm ShareMode) IsSkipBusy() bool {
	return sm == SM_SKIP
}

// WaitTimeout gets the wait timeout.
func (sm ShareMode) Timeout() time.Duration {
	return time.Duration(sm)
}

// SafeLockRead adds "read" reference to a named item.
func SafeLockRead(name string, mode ShareMode) bool {
	return globalSafeItems.LockRead(name, mode)
}

// SafeUnlockRead removes "read" reference from a named item.
func SafeUnlockRead(name string) {
	globalSafeItems.UnlockRead(name)
}

// SafeLockWrite adds "write" reference to a named item.
func SafeLockWrite(name string, mode ShareMode) bool {
	return globalSafeItems.LockWrite(name, mode)
}

// SafeUnlockWrite removes "write" reference from a named item.
func SafeUnlockWrite(name string) {
	globalSafeItems.UnlockWrite(name)
}

// Safe item
type safeItem struct {
	name  string // name of the item
	wrefs int    // number of "write" references
	rrefs int    // number of "read" references
}

// Safe items
type safeItems struct {
	items map[string]*safeItem
	lock  sync.Mutex
}

// create new safe items
func newSafeItems() *safeItems {
	return &safeItems{items: make(map[string]*safeItem)}
}

var (
	globalSafeItems = newSafeItems()
)

// LockRead adds "read" reference to a named item.
func (si *safeItems) LockRead(name string, mode ShareMode) bool {
	si.lock.Lock()
	defer si.lock.Unlock()

	var item *safeItem
	if item = si.items[name]; item != nil {
		if item.wrefs > 0 {
			safeLog.WithField("name", name).Warnf("[%s]: name is busy for reading (writers: %d)", SAFE, item.wrefs)
			return false // failed, BUSY!
		}
		item.rrefs++
	} else {
		item = &safeItem{name, 0, 1}
		si.items[name] = item
	}

	safeLog.WithField("name", name).Infof("[%s]: read lock acquired (readers:%d)", SAFE, item.rrefs)
	return true // OK
}

// UnlockRead removes "read" reference from a named item.
func (si *safeItems) UnlockRead(name string) {
	si.lock.Lock()
	defer si.lock.Unlock()

	if item := si.items[name]; item != nil {
		item.rrefs--
		safeLog.WithField("name", name).Infof("[%s]: read lock released (readers:%d)", SAFE, item.rrefs)
		if (item.rrefs + item.wrefs) == 0 {
			delete(si.items, name)
			safeLog.WithField("name", name).Debugf("[%s]: lock deleted (no references)", SAFE)
		}
	}
}

// LockWrite adds "write" reference to a named item.
func (si *safeItems) LockWrite(name string, mode ShareMode) bool {
	si.lock.Lock()
	defer si.lock.Unlock()

	var item *safeItem
	if item = si.items[name]; item != nil {
		if item.rrefs > 0 {
			safeLog.WithField("name", name).Warnf("[%s]: name is busy for writing (readers:%d)", SAFE, item.rrefs)
			return false // failed, BUSY!
		}
		item.wrefs++
	} else {
		item = &safeItem{name, 1, 0}
		si.items[name] = item
	}

	safeLog.WithField("name", name).Infof("[%s]: write lock acquired (writers:%d)", SAFE, item.wrefs)
	return true // OK
}

// UnlockWrite removes "write" reference from a named item.
func (si *safeItems) UnlockWrite(name string) {
	si.lock.Lock()
	defer si.lock.Unlock()

	if item := si.items[name]; item != nil {
		item.wrefs--
		safeLog.WithField("name", name).Infof("[%s]: write lock released (writers:%d)", SAFE, item.wrefs)
		if (item.rrefs + item.wrefs) == 0 {
			delete(si.items, name)
			safeLog.WithField("name", name).Debugf("[%s]: lock deleted (no references)", SAFE)
		}
	}
}
