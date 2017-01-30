package search

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test Result
func TestResultSimple(t *testing.T) {
	res := NewResult()
	assert.NotNil(t, res)
	assert.NotNil(t, res.ErrorChan)
	assert.EqualValues(t, 0, res.ErrorsReported())
	assert.NotNil(t, res.RecordChan)
	assert.EqualValues(t, 0, res.RecordsReported())
	assert.NotNil(t, res.DoneChan)
	assert.False(t, res.IsDone())
	assert.NotNil(t, res.CancelChan)
	assert.False(t, res.IsCancelled())
	assert.Equal(t, "Result{records:0, errors:0, done:false, cancelled:false, no stat}", res.String())

	// assign statistics
	res.Stat = NewStat("localhost")
	assert.Equal(t, "Result{records:0, errors:0, done:false, cancelled:false, stat:Stat{0 matches on 0 bytes in 0 ms (fabric: 0 ms), details:[], host:\"localhost\"}}", res.String())

	// report errors
	res.ReportError(nil)
	assert.EqualValues(t, 1, res.ErrorsReported())
	assert.Nil(t, <-res.ErrorChan)

	// report records
	res.ReportRecord(nil)
	res.ReportRecord(nil)
	assert.EqualValues(t, 2, res.RecordsReported())
	assert.Nil(t, <-res.RecordChan)
	assert.Nil(t, <-res.RecordChan)

	// simulate records reporting
	go func() {
		defer res.Close()
		defer res.ReportDone()

		for i := 0; i < 100; i++ {
			res.ReportError(fmt.Errorf("error-%d", i))
			res.ReportRecord(NewRecord(nil, nil))
		}

		for i := 0; i < 100000; i++ {
			res.ReportRecord(NewRecord(nil, nil))
		}

		for i := 0; i < 100000; i++ {
			res.ReportError(fmt.Errorf("error-%d", i))
		}

		for i := 0; i < 10000; i++ {
			res.ReportError(fmt.Errorf("error-%d", i))
			res.ReportRecord(NewRecord(nil, nil))
		}
	}()

	errors, records := res.Cancel()
	assert.EqualValues(t, 110100, errors)
	assert.EqualValues(t, 110100, records)
	assert.True(t, res.IsCancelled())
	assert.True(t, res.IsDone())

	// second Cancel does nothing
	errors, records = res.Cancel()
	assert.EqualValues(t, 0, errors)
	assert.EqualValues(t, 0, records)

	// as second ReportDone
	res.ReportDone()
	assert.True(t, res.IsDone())
	assert.Nil(t, <-res.RecordChan)
	assert.Nil(t, <-res.ErrorChan)

	// second Close() will panic
	// res.Close()
}
