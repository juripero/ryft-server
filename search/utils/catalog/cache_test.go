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
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
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
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// simple cache tests
func TestCacheUsual(t *testing.T) {
	SetLogLevelString(testLogLevel)

	cc := NewCache()
	if !assert.NotNil(t, cc) {
		return
	}
	cc.DropTimeout = 100 * time.Millisecond

	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	assert.NoError(t, os.MkdirAll(root, 0755))
	defer os.RemoveAll(root)

	assert.Nil(t, cc.Get(filepath.Join(root, "foo.catalog")))
	assert.False(t, cc.Drop(filepath.Join(root, "foo.catalog")))

	// open/close without cache
	log.Debugf("[test]: check simple open/close")
	cat, err := OpenCatalogNoCache(filepath.Join(root, "foo.catalog"))
	if assert.NoError(t, err) && assert.NotNil(t, cat) {
		assert.EqualValues(t, 0, cat.cacheRef)
		assert.NoError(t, cat.Close())
		assert.EqualValues(t, 0, cat.cacheRef)
		assert.Nil(t, cat.db) // DB is closed
		assert.NoError(t, cat.Close())
	}

	// open and get a few copies from cache
	log.Debugf("[test]: check open/close and cache put/get/drop")
	cat, err = OpenCatalogNoCache(filepath.Join(root, "foo.catalog"))
	if assert.NoError(t, err) && assert.NotNil(t, cat) {
		assert.EqualValues(t, 0, cat.cacheRef)
		cc.Put(cat.GetPath(), cat)
		assert.EqualValues(t, 1, cat.cacheRef)
		assert.Panics(t, func() {
			// failed to add second time!
			cc.Put(cat.GetPath(), cat)
		})

		c1 := cc.Get(cat.GetPath()) // ref++
		c2 := cc.Get(cat.GetPath()) // ref++
		if assert.True(t, cat == c1) && assert.True(t, cat == c2) {
			assert.EqualValues(t, 3, cat.cacheRef)
			assert.NoError(t, c1.Close()) // ref--
			assert.NoError(t, c2.Close()) // ref--
			assert.EqualValues(t, 1, cat.cacheRef)
		}

		assert.NoError(t, cat.Close()) // ref--
		assert.EqualValues(t, 0, cat.cacheRef)
		assert.True(t, cc.Drop(cat.GetPath()))
		assert.EqualValues(t, 0, cat.cacheRef)

		assert.NotNil(t, cat.db) // DB is NOT closed
		assert.NoError(t, cat.Close())
		assert.EqualValues(t, 0, cat.cacheRef)
		assert.Nil(t, cat.db) // DB is closed

		assert.Nil(t, cc.Get(filepath.Join(root, "foo.catalog")))
	}

	// open and get a few copies from cache (check drop timeout)
	log.Debugf("[test]: check open/close and cache put/get/drop by timeout")
	cat, err = OpenCatalogNoCache(filepath.Join(root, "foo.catalog"))
	if assert.NoError(t, err) && assert.NotNil(t, cat) {
		assert.EqualValues(t, 0, cat.cacheRef)
		cc.Put(cat.GetPath(), cat)
		assert.EqualValues(t, 1, cat.cacheRef)

		c1 := cc.Get(cat.GetPath()) // ref++
		c2 := cc.Get(cat.GetPath()) // ref++
		if assert.True(t, cat == c1) && assert.True(t, cat == c2) {
			assert.EqualValues(t, 3, cat.cacheRef)
			assert.NoError(t, c1.Close()) // ref--
			assert.NoError(t, c2.Close()) // ref--
			assert.EqualValues(t, 1, cat.cacheRef)
		}

		assert.NoError(t, cat.Close()) // ref--
		assert.EqualValues(t, 0, cat.cacheRef)
		// drop timer started here...

		// wait a bit and refresh drop-timer
		time.Sleep(cc.DropTimeout / 2)
		c3 := cc.Get(cat.GetPath()) // ref++
		if assert.True(t, cat == c3) {
			assert.EqualValues(t, 1, cat.cacheRef)
			assert.NoError(t, c3.Close()) // ref--
			assert.EqualValues(t, 0, cat.cacheRef)
			time.Sleep(cc.DropTimeout / 2)
			assert.NoError(t, c3.Close()) // ref--
			assert.EqualValues(t, -1, cat.cacheRef)
		}

		// wait drop timeout...
		time.Sleep(cc.DropTimeout + 200*time.Millisecond)
		assert.EqualValues(t, 0, cat.cacheRef)
		assert.Nil(t, cat.db) // DB is closed

		assert.Nil(t, cc.Get(filepath.Join(root, "foo.catalog")))
	}

	assert.Empty(t, cc.cached)
}
