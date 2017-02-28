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
	"sync"
)

// Safe item
type safeItem struct {
	name  string // name of the item
	wrefs int    // number of "write" references
	rrefs int    // number of "read" references
}

var (
	safeItems     = make(map[string]*safeItem)
	safeItemsLock = sync.Mutex{}
)

// SafeLockRead adds "read" reference to a named item.
func SafeLockRead(name string) bool {
	safeItemsLock.Lock()
	defer safeItemsLock.Unlock()

	if item, ok := safeItems[name]; ok {
		if item.wrefs > 0 {
			return false // failed, BUSY!
		}
		item.rrefs++
	} else {
		safeItems[name] = &safeItem{name, 0, 1}
	}

	return true // OK
}

// SafeUnlockRead removes "read" reference from a named item.
func SafeUnlockRead(name string) {
	safeItemsLock.Lock()
	defer safeItemsLock.Unlock()

	if item, ok := safeItems[name]; ok {
		if item.rrefs--; (item.rrefs + item.wrefs) == 0 {
			delete(safeItems, name)
		}
	}
}

// SafeLockWrite adds "write" reference to a named item.
func SafeLockWrite(name string) bool {
	safeItemsLock.Lock()
	defer safeItemsLock.Unlock()

	if item, ok := safeItems[name]; ok {
		if item.rrefs > 0 {
			return false // failed, BUSY!
		}
		item.wrefs++
	} else {
		safeItems[name] = &safeItem{name, 1, 0}
	}

	return true // OK
}

// SafeUnlockWrite removes "write" reference from a named item.
func SafeUnlockWrite(name string) {
	safeItemsLock.Lock()
	defer safeItemsLock.Unlock()

	if item, ok := safeItems[name]; ok {
		if item.wrefs--; (item.rrefs + item.wrefs) == 0 {
			delete(safeItems, name)
		}
	}
}
