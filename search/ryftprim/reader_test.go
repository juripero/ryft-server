package ryftprim

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/stretchr/testify/assert"
)

// generate fake ryftprim content (3 records)
func testFakeRyftprim3(od, oi *os.File, delim string) {
	// first record
	od.WriteString("hello")
	od.WriteString(delim)
	//od.Flush()
	oi.WriteString("/ryftone/1.txt,100,5,0\n")
	//oi.Flush()

	// second record
	oi.WriteString("2.txt,200,5,n/a\n") // FALLBACK to absolute
	//oi.Flush()
	time.Sleep(100 * time.Millisecond) // emulate "no data"
	od.WriteString("hello")
	od.WriteString(delim)
	//od.Flush()

	// third record
	od.WriteString("hello")
	od.WriteString(delim)
	//od.Flush()
	time.Sleep(100 * time.Millisecond)
	oi.WriteString("/ryftone/3.txt,300,5") // first INDEX part
	//oi.Flush()
	time.Sleep(100 * time.Millisecond)
	oi.WriteString(",1\n") // second INDEX part
	//oi.Flush()
}

// get reader's fake paths
func testReaderFake() (index, data, delim string) {
	index = fmt.Sprintf("/tmp/ryftprim-%x-index.txt", time.Now().UnixNano())
	data = fmt.Sprintf("/tmp/ryfptrim-%x-data.bin", time.Now().UnixNano())
	delim = "\r\n\f"
	return
}

// valid results
func TestReaderUsual(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		od, _ := os.Create(dataPath)
		oi, _ := os.Create(indexPath)
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		testFakeRyftprim3(od, oi, delimiter)

		// soft stop
		time.Sleep(100 * time.Millisecond)
		rr.stop()
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 0, res.ErrorsReported()) &&
		assert.EqualValues(t, 3, res.RecordsReported()) {
		assert.EqualValues(t, 3*(5+len(delimiter)), rr.totalDataLength)

		// check first record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "1.txt", rec.Index.File)
			assert.EqualValues(t, 100, rec.Index.Offset)
			assert.EqualValues(t, 5, rec.Index.Length)
			assert.EqualValues(t, 0, rec.Index.Fuzziness)
			assert.EqualValues(t, "hello", rec.RawData)
		}

		// check second record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "2.txt", rec.Index.File)
			assert.EqualValues(t, 200, rec.Index.Offset)
			assert.EqualValues(t, 5, rec.Index.Length)
			assert.EqualValues(t, -1, rec.Index.Fuzziness)
			assert.EqualValues(t, "hello", rec.RawData)
		}

		// check third record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "3.txt", rec.Index.File)
			assert.EqualValues(t, 300, rec.Index.Offset)
			assert.EqualValues(t, 5, rec.Index.Length)
			assert.EqualValues(t, 1, rec.Index.Fuzziness)
			assert.EqualValues(t, "hello", rec.RawData)
		}
	}
}

// valid results (no data read)
func TestReaderNoData(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = false // !!! NO DATA

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		od, _ := os.Create(dataPath)
		oi, _ := os.Create(indexPath)
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		testFakeRyftprim3(od, oi, delimiter)

		// soft stop
		time.Sleep(100 * time.Millisecond)
		rr.stop()
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 0, res.ErrorsReported()) &&
		assert.EqualValues(t, 3, res.RecordsReported()) {
		assert.EqualValues(t, 3*(5+len(delimiter)), rr.totalDataLength)

		// check first record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "1.txt", rec.Index.File)
			assert.EqualValues(t, 100, rec.Index.Offset)
			assert.EqualValues(t, 5, rec.Index.Length)
			assert.EqualValues(t, 0, rec.Index.Fuzziness)
			assert.Nil(t, rec.RawData) // assert.EqualValues(t, "hello", rec.RawData)
		}

		// check second record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "2.txt", rec.Index.File)
			assert.EqualValues(t, 200, rec.Index.Offset)
			assert.EqualValues(t, 5, rec.Index.Length)
			assert.EqualValues(t, -1, rec.Index.Fuzziness)
			assert.Nil(t, rec.RawData) // assert.EqualValues(t, "hello", rec.RawData)
		}

		// check third record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "3.txt", rec.Index.File)
			assert.EqualValues(t, 300, rec.Index.Offset)
			assert.EqualValues(t, 5, rec.Index.Length)
			assert.EqualValues(t, 1, rec.Index.Fuzziness)
			assert.Nil(t, rec.RawData) // assert.EqualValues(t, "hello", rec.RawData)
		}
	}
}

// valid results with limit
func TestReaderLimit(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true
	rr.Limit = 2 // !!! only TWO records expected

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		od, _ := os.Create(dataPath)
		oi, _ := os.Create(indexPath)
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		testFakeRyftprim3(od, oi, delimiter)

		// soft stop
		time.Sleep(100 * time.Millisecond)
		rr.stop()
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 0, res.ErrorsReported()) &&
		assert.EqualValues(t, 2, res.RecordsReported()) {
		assert.EqualValues(t, 2*(5+len(delimiter)), rr.totalDataLength)

		// check first record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "1.txt", rec.Index.File)
			assert.EqualValues(t, 100, rec.Index.Offset)
			assert.EqualValues(t, 5, rec.Index.Length)
			assert.EqualValues(t, 0, rec.Index.Fuzziness)
			assert.EqualValues(t, "hello", rec.RawData)
		}

		// check second record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "2.txt", rec.Index.File)
			assert.EqualValues(t, 200, rec.Index.Offset)
			assert.EqualValues(t, 5, rec.Index.Length)
			assert.EqualValues(t, -1, rec.Index.Fuzziness)
			assert.EqualValues(t, "hello", rec.RawData)
		}
	}
}

// bad results (failed to open INDEX file)
func TestReaderFailedToOpenIndex(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		od, _ := os.OpenFile(dataPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_EXCL, 0222)  // WRITE-ONLY
		oi, _ := os.OpenFile(indexPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_EXCL, 0222) // WRITE-ONLY
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		testFakeRyftprim3(od, oi, delimiter)
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 1, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {

		// check first error
		if err := <-res.ErrorChan; assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "failed to open INDEX file")
		}
	}
}

// bad results (failed to open DATA file)
func TestReaderFailedToOpenData(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		od, _ := os.OpenFile(dataPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_EXCL, 0222) // WRITE-ONLY
		oi, _ := os.Create(indexPath)
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		testFakeRyftprim3(od, oi, delimiter)
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 1, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {

		// check first error
		if err := <-res.ErrorChan; assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "failed to open DATA file")
		}
	}
}

// bad results (cancel to open INDEX file)
func TestReaderCancelToOpenIndex(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		rr.cancel()
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 0, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {
		assert.True(t, rr.isCancelled() != 0)
		assert.True(t, rr.isStopped() != 0)
	}
}

// bad results (failed to read INDEX)
func TestReaderFailedToReadIndex(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		od, _ := os.Create(dataPath)
		oi, _ := os.Create(indexPath)
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		oi.WriteString("1.txt,100") // no ",5,0\n"
		//oi.Flush()

		time.Sleep(1500 * time.Millisecond) // should be greater than read*limit
		rr.finish()                         // no STOP, no CANCEL
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 1, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {
		// check first error
		if err := <-res.ErrorChan; assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "cancelled by attempt limit")
		}
	}
}

// bad results (cancel to read INDEX)
func TestReaderCancelToReadIndex(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		od, _ := os.Create(dataPath)
		oi, _ := os.Create(indexPath)
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		oi.WriteString("1.txt,100") // no ",5,0\n"
		//oi.Flush()

		time.Sleep(100 * time.Millisecond)
		rr.cancel() // cancel instead of stop
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 0, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {
		assert.True(t, rr.isCancelled() != 0)
		assert.True(t, rr.isStopped() != 0)
	}
}

// bad results (cancel to open DATA file)
func TestReaderCancelToOpenData(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		//od, _ := os.OpenFile(dataPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_EXCL, 0222) // WRITE-ONLY
		oi, _ := os.Create(indexPath)
		//assert.NotNil(t, od)
		assert.NotNil(t, oi)
		//defer od.Close()
		defer oi.Close()

		oi.WriteString("1.txt,100,5,0\n")
		//oi.Flush()

		time.Sleep(100 * time.Millisecond)
		rr.cancel()
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 0, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {
		assert.True(t, rr.isCancelled() != 0)
		assert.True(t, rr.isStopped() != 0)
	}
}

// bad results (failed to parse INDEX)
func TestReaderFailedToParseIndex(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		od, _ := os.Create(dataPath)
		oi, _ := os.Create(indexPath)
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		testFakeRyftprim3(od, oi, delimiter)

		// fourth record
		od.WriteString("hello")
		od.WriteString(delimiter)
		//od.Flush()
		time.Sleep(100 * time.Millisecond)
		oi.WriteString("4.txt,300,5\n") // no FUZZINESS
		//oi.Flush()
		time.Sleep(100 * time.Millisecond)

		// soft stop
		time.Sleep(100 * time.Millisecond)
		rr.stop()
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 1, res.ErrorsReported()) &&
		assert.EqualValues(t, 3, res.RecordsReported()) {
		// check first error
		if err := <-res.ErrorChan; assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "failed to parse INDEX")
		}
	}
}

// bad results (failed to read DATA)
func TestReaderFailedToReadData(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		od, _ := os.Create(dataPath)
		oi, _ := os.Create(indexPath)
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		oi.WriteString("1.txt,100,5,0\n")
		//oi.Flush()
		od.WriteString("hell") // no "o" and no delimiter
		//od.Flush()

		time.Sleep(100 * time.Millisecond)
		rr.stop()
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 1, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {
		// check first error
		if err := <-res.ErrorChan; assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "failed to read DATA")
		}
	}
}

// bad results (cancel to read DATA)
func TestReaderCancelToReadData(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		od, _ := os.Create(dataPath)
		oi, _ := os.Create(indexPath)
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		oi.WriteString("1.txt,100,5,0\n")
		//oi.Flush()
		od.WriteString("hell") // no "o" and no delimiter
		//od.Flush()

		time.Sleep(100 * time.Millisecond)
		rr.cancel() // cancel instead of stop
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 0, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {
		assert.True(t, rr.isCancelled() != 0)
		assert.True(t, rr.isStopped() != 0)
	}
}

// bad results (failed to read DATA delimiter)
func TestReaderFailedToReadDelim(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		od, _ := os.Create(dataPath)
		oi, _ := os.Create(indexPath)
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		oi.WriteString("1.txt,100,5,0\n")
		//oi.Flush()
		od.WriteString("hello")
		od.WriteString("\r") // no "\n\f"
		//od.Flush()

		time.Sleep(100 * time.Millisecond)
		rr.stop()
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 1, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {
		// check first error
		if err := <-res.ErrorChan; assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "failed to read DATA delimiter")
		}
	}
}

// bad results (unexpected DATA delimiter)
func TestReaderUnexpectedDelim(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		od, _ := os.Create(dataPath)
		oi, _ := os.Create(indexPath)
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		oi.WriteString("1.txt,100,5,0\n")
		//oi.Flush()
		od.WriteString("hello")
		od.WriteString("\f\n\r") // unexpected
		//od.Flush()

		time.Sleep(100 * time.Millisecond)
		rr.stop()
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 1, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {
		// check first error
		if err := <-res.ErrorChan; assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), "unexpected delimiter found at")
		}
	}
}

// bad results (cancel to read DATA delimiter)
func TestReaderCancelToReadDelim(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second) // initial delay

		od, _ := os.Create(dataPath)
		oi, _ := os.Create(indexPath)
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		oi.WriteString("1.txt,100,5,0\n")
		//oi.Flush()
		od.WriteString("hello")
		od.WriteString("\r") // no "\n\f"
		//od.Flush()

		time.Sleep(100 * time.Millisecond)
		rr.cancel() // cancel instead of stop
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 0, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {
		assert.True(t, rr.isCancelled() != 0)
		assert.True(t, rr.isStopped() != 0)
	}
}
