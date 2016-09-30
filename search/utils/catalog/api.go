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
	"errors"
	"log"
	"os"
)

// global cache instance
var globalCache = NewCache()

// ErrNotCatalog is used to indicate the file is not a catalog meta-data file.
var ErrNotCatalog = errors.New("not a catalog")

// IsCatalog check if file is a catalog
func IsCatalog(path string) bool {
	if s, err := os.Stat(path); os.IsNotExist(err) {
		return false // not exist
	} else if s.Size() <= 0 {
		return false // bad size
	}

	cat, err := OpenCatalog(path, true)
	if err != nil {
		return false
	}
	defer cat.Close()

	if ok, err := cat.checkSchemeSync(); err != nil || !ok {
		return false
	}

	return true
}

// OpenCatalog opens catalog file.
func OpenCatalog(path string, readOnly bool) (*Catalog, error) {
	cat, cached, err := getCatalog(path)
	if err != nil {
		return nil, err
	}

	// update database scheme
	if !cached && !readOnly {
		log.Printf("updating catalog scheme: %s", path)
		if err := cat.updateSchemeSync(); err != nil {
			cat.Close()
			return nil, err
		}
	}

	return cat, nil // OK
}

// get catalog (from cache or new)
func getCatalog(path string) (*Catalog, bool, error) {
	globalCache.Lock()
	defer globalCache.Unlock()

	// try to get existing catalog
	if cat := globalCache.get(path); cat != nil {
		return cat, true, nil // OK
	}

	// create new one and put to cache
	cat, err := openCatalog(path)
	if err == nil && cat != nil {
		globalCache.put(path, cat)
		cat.cacheAddRef()
	}

	return cat, false, err
}
