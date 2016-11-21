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

package catalog

import (
	"sync"
	"time"
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
		cat.cacheAddRef()
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
	cat.CacheDropTimeout = cc.DropTimeout
	cat.cache = cc

	cc.cached[path] = cat
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
		return true
	}

	return false // does not exist
}
