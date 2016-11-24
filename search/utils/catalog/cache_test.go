package catalog

import (
	"os"
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

	os.MkdirAll("/tmp/ryft/", 0755)
	defer os.RemoveAll("/tmp/ryft/")
	assert.Nil(t, cc.Get("/tmp/ryft/foo.catalog"))
	assert.False(t, cc.Drop("/tmp/ryft/foo.catalog"))

	// open/close without cache
	log.Debugf("[test]: check simple open/close")
	cat, err := OpenCatalogNoCache("/tmp/ryft/foo.catalog")
	if assert.NoError(t, err) && assert.NotNil(t, cat) {
		assert.EqualValues(t, 0, cat.cacheRef)
		assert.NoError(t, cat.Close())
		assert.EqualValues(t, 0, cat.cacheRef)
		assert.Nil(t, cat.db) // DB is closed
		assert.NoError(t, cat.Close())
	}

	// open and get a few copies from cache
	log.Debugf("[test]: check open/close and cache put/get/drop")
	cat, err = OpenCatalogNoCache("/tmp/ryft/foo.catalog")
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

		assert.Nil(t, cc.Get("/tmp/ryft/foo.catalog"))
	}

	// open and get a few copies from cache (check drop timeout)
	log.Debugf("[test]: check open/close and cache put/get/drop by timeout")
	cat, err = OpenCatalogNoCache("/tmp/ryft/foo.catalog")
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

		assert.Nil(t, cc.Get("/tmp/ryft/foo.catalog"))
	}

	assert.Empty(t, cc.cached)
}
