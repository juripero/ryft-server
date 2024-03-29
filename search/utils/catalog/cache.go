/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2018, Ryft Systems, Inc.
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

package catalog

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
)

// global cache instance
var globalCache = NewCache()

// Cache contains list of cached catalogs
type Cache struct {
	DropTimeout time.Duration

	cached map[string]*Catalog // db path -> Catalog cache

	sync.Mutex
}

// NewCache creates new (empty cache)
func NewCache() *Cache {
	cc := new(Cache)
	cc.DropTimeout = DefaultCacheDropTimeout
	cc.cached = make(map[string]*Catalog)
	return cc
}

// Get gets existing catalog from cache.
// return `nil` if no catalog found!
func (cc *Cache) Get(path string) *Catalog {
	cc.Lock()
	defer cc.Unlock()

	return cc.get(path)
}

// Get gets existing catalog from cache (unsynchronized).
func (cc *Cache) get(path string) *Catalog {
	// try to get existing catalog
	if cat, ok := cc.cached[path]; ok && cat != nil {
		cc.cacheAddRef(cat)
		return cat
	}

	return nil // not found
}

// Put saves catalog to cache.
func (cc *Cache) Put(path string, cat *Catalog) {
	cc.Lock()
	defer cc.Unlock()

	cc.put(path, cat)
}

// Put saves catalog to cache (unsynchronized).
func (cc *Cache) put(path string, cat *Catalog) {
	if cat.cache != nil {
		panic("catalog is already used")
	}

	cc.cached[path] = cat
	cc.cacheAddRef(cat)
	cat.cache = cc
}

// Drop removes the catalog from cache.
func (cc *Cache) Drop(path string) bool {
	cc.Lock()
	defer cc.Unlock()

	return cc.drop(path)
}

// Drop removes the catalog from cache (unsynchronized).
func (cc *Cache) drop(path string) bool {
	// try to drop existing catalog
	if cat, ok := cc.cached[path]; ok && cat != nil {
		delete(cc.cached, path)
		cat.cache = nil

		atomic.StoreInt32(&cat.cacheRef, 0)
		cc.stopDropTimerSync(cat)
		return true
	}

	return false // does not exist
}

// DropFromCache removes all the catalogs starting with path.
func DropFromCache(path string) int {
	return globalCache.DropAll(path)
}

// DropAll removes all the catalogs starting with path.
func (cc *Cache) DropAll(path string) int {
	cc.Lock()
	defer cc.Unlock()

	return cc.dropAll(path)
}

// dropAll removes all the catalogs starting with path (unsynchronized).
func (cc *Cache) dropAll(path string) int {
	n := 0

	// try to drop existing catalogs
	for cpath, cat := range cc.cached {
		if search.IsRelativeToHome(path, cpath) {
			cat.log().Debugf("[%s]: *** drop from cache", TAG)
			delete(cc.cached, cpath)
			cat.cache = nil

			atomic.StoreInt32(&cat.cacheRef, 0)
			cc.stopDropTimerSync(cat)
			n += 1
		}
	}

	return n
}

// release catalog reference
func (cc *Cache) release(cat *Catalog) {
	cc.cacheRelease(cat)
}

// cache: add reference
func (cc *Cache) cacheAddRef(cat *Catalog) {
	ref := atomic.AddInt32(&cat.cacheRef, +1)
	//cat.log().WithField("ref", ref).Debugf("[%s]: reference++", TAG) // FIXME: DEBUG
	if ref == 1 {
		cc.stopDropTimerSync(cat) // just in case
	}
}

// cache: release
func (cc *Cache) cacheRelease(cat *Catalog) {
	ref := atomic.AddInt32(&cat.cacheRef, -1)
	//cat.log().WithField("ref", ref).Debugf("[%s]: reference--", TAG) // FIXME: DEBUG
	if ref <= 0 {
		// for the last reference start the drop timer
		cc.startDropTimerSync(cat, cc.DropTimeout)
	}
}

// start drop timer (synchronized)
func (cc *Cache) startDropTimerSync(cat *Catalog, timeout time.Duration) {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	if cat.cacheDrop != nil && cat.cacheDrop.Reset(timeout) {
		// timer is updated
		cat.log().Debugf("[%s]: refresh drop-timer", TAG) // FIXME: DEBUG
	} else {
		cat.cacheDrop = time.AfterFunc(timeout, func() {
			cat.log().Debugf("[%s]: drop-timeout is expired...", TAG)
			if cat.DropFromCache() {
				cat.Close()
			}
		})
		cat.log().WithField("timeout", timeout).Debugf("[%s]: start drop-timer", TAG) // FIXME: DEBUG
	}
}

// stop drop timer if any (synchronized)
func (cc *Cache) stopDropTimerSync(cat *Catalog) {
	cat.mutex.Lock()
	defer cat.mutex.Unlock()

	if cd := cat.cacheDrop; cd != nil {
		cat.cacheDrop = nil
		if cd.Stop() {
			cat.log().Debugf("[%s]: stop drop-timer", TAG) // FIXME: DEBUG
		}
	}
}
