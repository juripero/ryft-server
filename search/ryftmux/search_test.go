package ryftmux

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/stretchr/testify/assert"
)

// Run asynchronous "/search" or "/count" operation.
func (fe *fakeEngine) Search(cfg *search.Config) (*search.Result, error) {
	if fe.ErrorForSearch != nil {
		return nil, fe.ErrorForSearch
	}

	res := search.NewResult()
	go func() {
		defer res.Close()
		defer res.ReportDone()

		// report fake data
		nr := int64(fe.RecordsToReport)
		ne := int64(fe.ErrorsToReport)
		cancelled := 0
		for (nr > 0 || ne > 0) && cancelled < 10 {
			if rand.Int63n(ne+nr) >= ne {
				idx := search.NewIndex(fmt.Sprintf("file-%d.txt", nr), uint64(nr), uint64(nr))
				idx.UpdateHost(fe.Host)
				data := []byte(fmt.Sprintf("data-%d", nr))
				rec := search.NewRecord(idx, data)
				res.ReportRecord(rec)
				nr--
			} else {
				err := fmt.Errorf("error-%d", ne)
				res.ReportError(err)
				ne--
			}

			if res.IsCancelled() {
				cancelled++ // emulate cancel delay here
			}

			if fe.ReportLatency > 0 {
				time.Sleep(fe.ReportLatency)
			}
		}

		res.Stat = search.NewStat(fe.Host)
		res.Stat.Matches = uint64(fe.RecordsToReport)
		res.Stat.TotalBytes = uint64(rand.Int63n(1000000000) + 1)
		res.Stat.Duration = uint64(rand.Int63n(1000) + 1)
		res.Stat.FabricDuration = res.Stat.Duration / 2
	}()

	return res, nil // OK for now
}

// drain the results
func drain(res *search.Result) (records int, errors int) {
	for {
		select {
		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				errors++
			}

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				records++
			}

		case <-res.DoneChan:
			for _ = range res.ErrorChan {
				errors++
			}
			for _ = range res.RecordChan {
				records++
			}

			return
		}
	}
}

// Check multiplexing of search results
func TestEngineSearchUsual(t *testing.T) {
	SetLogLevelString(testLogLevel)

	f1 := newFake(100000, 100)
	f1.Host = "host-1"

	f2 := newFake(1000, 10)
	f2.Host = "host-2"

	f3 := newFake(10, 1)
	f3.Host = "host-3"

	// valid (usual case)
	engine, err := NewEngine(f1, f2, f3)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello")

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			records, errors := drain(res)

			assert.EqualValues(t, f1.RecordsToReport+f2.RecordsToReport+f3.RecordsToReport, res.RecordsReported())
			assert.EqualValues(t, f1.ErrorsToReport+f2.ErrorsToReport+f3.ErrorsToReport, res.ErrorsReported())
			assert.EqualValues(t, f1.RecordsToReport+f2.RecordsToReport+f3.RecordsToReport, records)
			assert.EqualValues(t, f1.ErrorsToReport+f2.ErrorsToReport+f3.ErrorsToReport, errors)
		}
	}
}

// Check multiplexing of search results with limit.
func TestEngineSearchLimit(t *testing.T) {
	SetLogLevelString(testLogLevel)

	f1 := newFake(100000, 100)
	f1.Host = "host-1"

	f2 := newFake(1000, 10)
	f2.ReportLatency = time.Millisecond
	f2.Host = "host-2"

	f3 := newFake(10, 1)
	f3.ReportLatency = 10 * time.Millisecond
	f3.Host = "host-3"

	// valid (usual case)
	engine, err := NewEngine(f1, f2, f3)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello")
		cfg.Limit = 500

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			records, _ := drain(res)

			assert.EqualValues(t, cfg.Limit, res.RecordsReported())
			//assert.EqualValues(t, f1.ErrorsToReport+f2.ErrorsToReport+f3.ErrorsToReport, res.ErrorsReported())
			assert.EqualValues(t, cfg.Limit, records)
			//assert.EqualValues(t, f1.ErrorsToReport+f2.ErrorsToReport+f3.ErrorsToReport, errors)
		}
	}
}

// Check multiplexing of search results
// failed to do search on a backend
func TestEngineSearchFailed1(t *testing.T) {
	SetLogLevelString(testLogLevel)

	f1 := newFake(100000, 100)
	f1.Host = "host-1"
	f1.ErrorForSearch = fmt.Errorf("disabled")

	f2 := newFake(1000, 10)
	f2.Host = "host-2"

	f3 := newFake(10, 1)
	f3.Host = "host-3"

	engine, err := NewEngine(f1, f2, f3)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello")

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			records, errors := drain(res)

			assert.EqualValues(t /*f1.RecordsToReport*/, 0+f2.RecordsToReport+f3.RecordsToReport, res.RecordsReported())
			assert.EqualValues(t /*f1.ErrorsToReport*/, 1+f2.ErrorsToReport+f3.ErrorsToReport, res.ErrorsReported())
			assert.EqualValues(t /*f1.RecordsToReport*/, 0+f2.RecordsToReport+f3.RecordsToReport, records)
			assert.EqualValues(t /*f1.ErrorsToReport*/, 1+f2.ErrorsToReport+f3.ErrorsToReport, errors)
		}
	}
}

// Check multiplexing of search results with cancel.
func TestEngineSearchCancel(t *testing.T) {
	SetLogLevelString(testLogLevel)

	f1 := newFake(100000, 100)
	f1.ReportLatency = time.Millisecond
	f1.Host = "host-1"

	f2 := newFake(1000, 10)
	f2.ReportLatency = time.Millisecond
	f2.Host = "host-2"

	f3 := newFake(10, 1)
	f3.Host = "host-3"

	// valid (usual case)
	engine, err := NewEngine(f1, f2, f3)
	if assert.NoError(t, err) && assert.NotNil(t, engine) {
		cfg := search.NewConfig("hello")

		res, err := engine.Search(cfg)
		if assert.NoError(t, err) && assert.NotNil(t, res) {
			go func() {
				time.Sleep(200 * time.Millisecond)
				res.Cancel() // cancel all
			}()

			_, _ = drain(res)

			assert.True(t, res.IsCancelled())
		}
	}
}
