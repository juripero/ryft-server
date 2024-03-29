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

package ryfthttp

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	json_codec "github.com/getryft/ryft-server/rest/codec/json"
	codec "github.com/getryft/ryft-server/rest/codec/msgpack.v1"
	format "github.com/getryft/ryft-server/rest/format/raw"
	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/testfake"
	"github.com/stretchr/testify/assert"
)

// set test log level
func testSetLogLevel() {
	SetLogLevelString(testLogLevel)
	testfake.SetLogLevelString(testLogLevel)
}

// do fake GET /search
func (fs *fakeServer) doSearch(w http.ResponseWriter, req *http.Request) {
	// report fake data
	nr := int64(fs.RecordsToReport)
	ne := int64(fs.ErrorsToReport)
	cancelled := 0

	enc, _ := codec.NewStreamEncoder(w)
	defer enc.Close()

	w.Header().Set("Content-Type", codec.MIME)
	w.WriteHeader(http.StatusOK)

	if fs.BadTagCase {
		w.Write([]byte("\x81\xa3\x62\x61\x64\xc3")) // msgpack: {"bad":true}
	}

	if fs.BadUnkTagCase {
		w.Write([]byte{0x7f}) // msgpack: 127
	}

	for (nr > 0 || ne > 0) && cancelled < 10 {
		if rand.Int63n(ne+nr) >= ne {
			idx := search.NewIndex(fmt.Sprintf("file-%d.txt", nr), uint64(nr), uint64(nr))
			idx.UpdateHost(fs.Host)
			data := []byte(fmt.Sprintf("data-%d", nr))
			rec := search.NewRecord(idx, data)

			xrec := format.FromRecord(rec)
			if fs.BadRecordCase {
				w.Write([]byte{0x01})               // TAG_REC
				w.Write([]byte("\xa3\x62\x61\x64")) // msgpack: "bad"
			} else {
				enc.EncodeRecord(xrec)
			}
			nr--
		} else {
			err := fmt.Errorf("error-%d", ne)
			if fs.BadErrorCase {
				w.Write([]byte{0x02})                       // TAG_ERR
				w.Write([]byte("\x81\xa3\x62\x61\x64\xc3")) // msgpack: {"bad":true}
			} else {
				enc.EncodeError(err)
			}
			ne--
		}

		//if res.IsCancelled() {
		//	cancelled++ // emulate cancel delay here
		//}

		if fs.ReportLatency > 0 {
			time.Sleep(fs.ReportLatency)
		}
	}

	stat := search.NewStat(fs.Host)
	stat.Matches = uint64(fs.RecordsToReport)
	stat.TotalBytes = uint64(rand.Int63n(1000000000) + 1)
	stat.Duration = uint64(rand.Int63n(1000) + 1)
	stat.FabricDuration = stat.Duration / 2

	xstat := format.FromStat(stat)
	if fs.BadStatCase {
		w.Write([]byte{0x03})               // TAG_STAT
		w.Write([]byte("\xa3\x62\x61\x64")) // msgpack: "bad"
	} else {
		enc.EncodeStat(xstat)
	}
}

// do fake GET /count
func (fs *fakeServer) doCount0(w http.ResponseWriter, req *http.Request) {
	//enc, _ := codec.NewStreamEncoder(w)
	//defer enc.Close()

	w.Header().Set("Content-Type", json_codec.MIME)
	w.WriteHeader(http.StatusOK)

	time.Sleep(fs.ReportLatency)

	stat := search.NewStat(fs.Host)
	stat.Matches = uint64(fs.RecordsToReport)
	stat.TotalBytes = uint64(rand.Int63n(1000000000) + 1)
	stat.Duration = uint64(rand.Int63n(1000) + 1)
	stat.FabricDuration = stat.Duration / 2

	xstat := format.FromStat(stat)
	enc := json.NewEncoder(w)
	enc.Encode(xstat)
}

// Check valid search results
func TestEngineSearchUsual(t *testing.T) {
	testSetLogLevel()

	fs := newFake(1000, 100)
	fs.Host = "host-1"

	go func() {
		err := fs.server.ListenAndServe()
		assert.NoError(t, err, "failed to start fake server")
	}()
	time.Sleep(100 * time.Millisecond) // wait a bit until server is started
	defer func() {
		fs.server.Stop(0)
		time.Sleep(100 * time.Millisecond) // wait a bit until server is stopped
	}()

	// valid (usual case)
	engine, err := NewEngine(map[string]interface{}{
		"server-url": fs.location(),
		"auth-token": "Basic: any-value-ignored",
		"local-only": true,
	})
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "1.txt")
		cfg.ReportIndex = true

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			records, errors := testfake.Drain(res)

			assert.EqualValues(t, fs.RecordsToReport, res.RecordsReported())
			assert.EqualValues(t, fs.ErrorsToReport, res.ErrorsReported())
			assert.EqualValues(t, fs.RecordsToReport, len(records))
			assert.EqualValues(t, fs.ErrorsToReport, len(errors))
		}
	}

	// bad case (failed to send request)
	oldUrl := engine.ServerURL
	engine.ServerURL = "bad-" + oldUrl
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "1.txt")
		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			<-res.DoneChan // wait results

			// log.Debugf("done, check results read")
			if assert.EqualValues(t, 1, res.ErrorsReported()) &&
				assert.EqualValues(t, 0, res.RecordsReported()) {

				// check first error
				if err := <-res.ErrorChan; assert.NotNil(t, err) {
					assert.Contains(t, err.Error(), "failed to send request")
				}
			}
		}
	}
	engine.ServerURL = oldUrl // restore back

	// bad case (invalid status)
	oldUrl = engine.ServerURL
	engine.ServerURL = oldUrl + "/bad"
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "1.txt")
		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			<-res.DoneChan // wait results

			// log.Debugf("done, check results read")
			if assert.EqualValues(t, 1, res.ErrorsReported()) &&
				assert.EqualValues(t, 0, res.RecordsReported()) {

				// check first error
				if err := <-res.ErrorChan; assert.NotNil(t, err) {
					assert.Contains(t, err.Error(), "invalid response status")
				}
			}
		}
	}
	engine.ServerURL = oldUrl // restore back
}

// Check valid count results
func TestEngineCountUsual(t *testing.T) {
	testSetLogLevel()

	fs := newFake(0, 0)
	fs.Host = "host-1"

	go func() {
		err := fs.server.ListenAndServe()
		assert.NoError(t, err, "failed to start fake server")
	}()
	time.Sleep(100 * time.Millisecond) // wait a bit until server is started
	defer func() {
		fs.server.Stop(0)
		time.Sleep(100 * time.Millisecond) // wait a bit until server is stopped
	}()

	// valid (usual case)
	engine, err := NewEngine(map[string]interface{}{
		"server-url": fs.location(),
		"auth-token": "Basic: any-value-ignored",
		"local-only": true,
	})
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "1.txt")
		cfg.ReportIndex = false

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			records, errors := testfake.Drain(res)

			assert.EqualValues(t, fs.RecordsToReport, res.RecordsReported())
			assert.EqualValues(t, fs.ErrorsToReport, res.ErrorsReported())
			assert.EqualValues(t, fs.RecordsToReport, len(records))
			assert.EqualValues(t, fs.ErrorsToReport, len(errors))
		}
	}
}

// Check bad search results
func TestEngineSearchFailedToDecode(t *testing.T) {
	testSetLogLevel()

	fs := newFake(1000, 100)
	fs.Host = "host-1"

	go func() {
		err := fs.server.ListenAndServe()
		assert.NoError(t, err, "failed to start fake server")
	}()
	time.Sleep(100 * time.Millisecond) // wait a bit until server is started
	defer func() {
		fs.server.Stop(0)
		time.Sleep(100 * time.Millisecond) // wait a bit until server is stopped
	}()

	engine, err := NewEngine(map[string]interface{}{
		"server-url": fs.location(),
		"auth-token": "Basic: any-value-ignored",
		"local-only": true,
	})

	// failed to decode tag
	fs.BadTagCase = true
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "1.txt")
		cfg.ReportIndex = true

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			<-res.DoneChan // wait results

			// log.Debugf("done, check results read")
			if assert.EqualValues(t, 1, res.ErrorsReported()) &&
				assert.EqualValues(t, 0, res.RecordsReported()) {

				// check first error
				if err := <-res.ErrorChan; assert.NotNil(t, err) {
					assert.Contains(t, err.Error(), "failed to decode next tag")
				}
			}
		}
	}
	fs.BadTagCase = false

	// failed to decode tag (unknown)
	fs.BadUnkTagCase = true
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "1.txt")
		cfg.ReportIndex = true

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			<-res.DoneChan // wait results

			// log.Debugf("done, check results read")
			if assert.EqualValues(t, 1, res.ErrorsReported()) &&
				assert.EqualValues(t, 0, res.RecordsReported()) {

				// check first error
				if err := <-res.ErrorChan; assert.NotNil(t, err) {
					assert.Contains(t, err.Error(), "unknown data tag received")
				}
			}
		}
	}
	fs.BadUnkTagCase = false

	// failed to decode record
	fs.BadRecordCase = true
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "1.txt")
		cfg.ReportIndex = true

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			<-res.DoneChan // wait results

			// log.Debugf("done, check results read")
			if assert.EqualValues(t, 1, res.ErrorsReported()) &&
				assert.EqualValues(t, 0, res.RecordsReported()) {

				// check first error
				if err := <-res.ErrorChan; assert.NotNil(t, err) {
					assert.Contains(t, err.Error(), "failed to decode record")
				}
			}
		}
	}
	fs.BadRecordCase = false

	// failed to decode error
	fs.BadErrorCase = true
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "1.txt")
		cfg.ReportIndex = true

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			<-res.DoneChan // wait results

			// log.Debugf("done, check results read")
			if assert.EqualValues(t, 1, res.ErrorsReported()) {

				// check first error
				if err := <-res.ErrorChan; assert.NotNil(t, err) {
					assert.Contains(t, err.Error(), "failed to decode error")
				}
			}
		}
	}
	fs.BadErrorCase = false

	// failed to decode statistics
	fs.BadStatCase = true
	fs.ErrorsToReport = 0
	if assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "1.txt")
		cfg.ReportIndex = true

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			<-res.DoneChan // wait results

			// log.Debugf("done, check results read")
			if assert.EqualValues(t, 1, res.ErrorsReported()) {

				// check first error
				if err := <-res.ErrorChan; assert.NotNil(t, err) {
					assert.Contains(t, err.Error(), "failed to decode statistics")
				}
			}
		}
	}
	fs.BadStatCase = false
}

// Check valid search results with cancel
func TestEngineSearchCancel(t *testing.T) {
	testSetLogLevel()

	fs := newFake(100000, 100)
	fs.ReportLatency = 100 * time.Millisecond
	fs.Host = "host-1"

	go func() {
		err := fs.server.ListenAndServe()
		assert.NoError(t, err, "failed to start fake server")
	}()
	time.Sleep(100 * time.Millisecond) // wait a bit until server is started
	defer func() {
		fs.server.Stop(0)
		time.Sleep(100 * time.Millisecond) // wait a bit until server is stopped
	}()

	// valid (usual case)
	engine, err := NewEngine(map[string]interface{}{
		"server-url": fs.location(),
		"auth-token": "Basic: any-value-ignored",
		"local-only": true,
	})
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello", "1.txt")
		cfg.ReportIndex = true

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			go func() {
				time.Sleep(1 * time.Second)
				log.Infof("[%s/test]: cancelling request!", TAG)
				res.Cancel() // cancel in 1 second
			}()
			_, _ = testfake.Drain(res)

			assert.True(t, res.IsCancelled())
			assert.True(t, res.IsDone())
		}
	}

	// the same for /count
	if assert.NotNil(t, engine) {
		fs.ReportLatency = 2 * time.Second
		cfg := search.NewConfig("hello", "1.txt")
		cfg.ReportIndex = false

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			go func() {
				time.Sleep(1 * time.Second)
				log.Infof("[%s/test]: cancelling request!", TAG)
				res.Cancel() // cancel in 1 second
			}()
			_, _ = testfake.Drain(res)

			assert.True(t, res.IsCancelled())
			assert.True(t, res.IsDone())
		}
	}
}
