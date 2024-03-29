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

package ryftprim

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils/view"
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

// generate fake ryftprim content (3 records) JSON array
func testFakeRyftprim3j(od, oi *os.File, delim string) {
	od.WriteString("[\n") // start of JSON array

	// first record
	od.WriteString(`{"text":"hello"}`)
	od.WriteString(delim)
	//od.Flush()
	oi.WriteString("/ryftone/1.txt,100,16,0\n")
	//oi.Flush()
	od.WriteString(",\n")

	// second record
	oi.WriteString("2.txt,200,16,n/a\n") // FALLBACK to absolute
	//oi.Flush()
	time.Sleep(100 * time.Millisecond) // emulate "no data"
	od.WriteString(`{"text":"hello"}`)
	od.WriteString(delim)
	od.WriteString(",\n")
	//od.Flush()

	// third record
	od.WriteString(`{"text":"hello"}`)
	od.WriteString(delim)
	//od.Flush()
	time.Sleep(100 * time.Millisecond)
	oi.WriteString("/ryftone/3.txt,300,16") // first INDEX part
	//oi.Flush()
	time.Sleep(100 * time.Millisecond)
	oi.WriteString(",1\n") // second INDEX part
	//oi.Flush()
	od.WriteString("\n]") // end of JSON array
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

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
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
		time.Sleep(200 * time.Millisecond) // initial delay

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
			assert.EqualValues(t, "{1.txt#100, len:5, d:0}", rec.Index.String())
			assert.EqualValues(t, "hello", rec.RawData)
		}

		// check second record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{2.txt#200, len:5, d:-1}", rec.Index.String())
			assert.EqualValues(t, "hello", rec.RawData)
		}

		// check third record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{3.txt#300, len:5, d:1}", rec.Index.String())
			assert.EqualValues(t, "hello", rec.RawData)
		}
	}
}

// valid results (JSON results)
func TestReaderJsonArrayUsual(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
	rr.CheckJsonArray = true
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
		time.Sleep(200 * time.Millisecond) // initial delay

		od, _ := os.Create(dataPath)
		oi, _ := os.Create(indexPath)
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		testFakeRyftprim3j(od, oi, delimiter)

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
		assert.EqualValues(t, 3*(2+16+len(delimiter))+2, rr.totalDataLength)

		// check first record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{1.txt#100, len:16, d:0}", rec.Index.String())
			assert.EqualValues(t, `{"text":"hello"}`, rec.RawData)
		}

		// check second record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{2.txt#200, len:16, d:-1}", rec.Index.String())
			assert.EqualValues(t, `{"text":"hello"}`, rec.RawData)
		}

		// check third record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{3.txt#300, len:16, d:1}", rec.Index.String())
			assert.EqualValues(t, `{"text":"hello"}`, rec.RawData)
		}
	}
}

// valid results + VIEW file
func TestReaderView(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath := "/tmp/ryftprim-index.txt"
	dataPath := "/tmp/ryfptrim-data.bin"
	viewPath := "/tmp/ryfptrim-view.bin"
	delimiter := "\r\n\f"

	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)
	defer os.RemoveAll(viewPath)

	rr := NewResultsReader(NewTask(nil, true), dataPath, indexPath, viewPath, delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true
	rr.MakeView = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(200 * time.Millisecond) // initial delay

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
			assert.EqualValues(t, "{1.txt#100, len:5, d:0}", rec.Index.String())
			assert.EqualValues(t, "hello", rec.RawData)
		}

		// check second record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{2.txt#200, len:5, d:-1}", rec.Index.String())
			assert.EqualValues(t, "hello", rec.RawData)
		}

		// check third record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{3.txt#300, len:5, d:1}", rec.Index.String())
			assert.EqualValues(t, "hello", rec.RawData)
		}
	}

	// check the VIEW file created
	if vf, err := view.Open(viewPath); assert.NoError(t, err) {
		if assert.EqualValues(t, 3, vf.Count()) {
			// first record
			if iBeg, iEnd, dBeg, dEnd, err := vf.Get(0); assert.NoError(t, err) {
				assert.EqualValues(t, []int64{0, 23, 0, 5}, []int64{iBeg, iEnd, dBeg, dEnd})
			}

			// second record
			if iBeg, iEnd, dBeg, dEnd, err := vf.Get(1); assert.NoError(t, err) {
				assert.EqualValues(t, []int64{23, 39, 8, 13}, []int64{iBeg, iEnd, dBeg, dEnd})
			}

			// third record
			if iBeg, iEnd, dBeg, dEnd, err := vf.Get(2); assert.NoError(t, err) {
				assert.EqualValues(t, []int64{39, 62, 16, 21}, []int64{iBeg, iEnd, dBeg, dEnd})
			}

			// failed
			if iBeg, iEnd, dBeg, dEnd, err := vf.Get(3); assert.Error(t, err) {
				assert.EqualValues(t, []int64{-1, -1, -1, -1}, []int64{iBeg, iEnd, dBeg, dEnd})
				assert.Contains(t, err.Error(), "VIEW out of range")
			}
		}
	}

	// read with VIEW file
	if true {
		rr := NewResultsReader(NewTask(nil, true), dataPath, indexPath, viewPath, delimiter)
		rr.RelativeToHome = "/ryftone"
		rr.OpenFilePollTimeout = 50 * time.Millisecond
		rr.ReadFilePollTimeout = 50 * time.Millisecond
		rr.ReadFilePollLimit = 20
		rr.ReadData = true
		rr.MakeView = false

		// emulate work:
		wg.Add(1)
		go func() {
			defer wg.Done()

			// soft stop
			time.Sleep(100 * time.Millisecond)
			rr.stop()
		}()

		res := search.NewResult()
		rr.process(res)
		wg.Wait()

		if assert.EqualValues(t, 0, res.ErrorsReported()) &&
			assert.EqualValues(t, 3, res.RecordsReported()) {

			// check first record
			if rec := <-res.RecordChan; assert.NotNil(t, rec) {
				assert.EqualValues(t, "{1.txt#100, len:5, d:0}", rec.Index.String())
				assert.EqualValues(t, "hello", rec.RawData)
			}

			// check second record
			if rec := <-res.RecordChan; assert.NotNil(t, rec) {
				assert.EqualValues(t, "{2.txt#200, len:5, d:-1}", rec.Index.String())
				assert.EqualValues(t, "hello", rec.RawData)
			}

			// check third record
			if rec := <-res.RecordChan; assert.NotNil(t, rec) {
				assert.EqualValues(t, "{3.txt#300, len:5, d:1}", rec.Index.String())
				assert.EqualValues(t, "hello", rec.RawData)
			}
		}
	}

	// read with VIEW file
	if true {
		rr := NewResultsReader(NewTask(nil, true), dataPath, indexPath, viewPath, delimiter)
		rr.RelativeToHome = "/ryftone"
		rr.OpenFilePollTimeout = 50 * time.Millisecond
		rr.ReadFilePollTimeout = 50 * time.Millisecond
		rr.ReadFilePollLimit = 20
		rr.ReadData = true
		rr.MakeView = false
		rr.Offset = 1
		rr.Limit = 1

		// emulate work:
		wg.Add(1)
		go func() {
			defer wg.Done()

			// soft stop
			time.Sleep(100 * time.Millisecond)
			rr.stop()
		}()

		res := search.NewResult()
		rr.process(res)
		wg.Wait()

		if assert.EqualValues(t, 0, res.ErrorsReported()) &&
			assert.EqualValues(t, 1, res.RecordsReported()) {

			// check first record
			/*if rec := <-res.RecordChan; assert.NotNil(t, rec) {
				assert.EqualValues(t, "{1.txt#100, len:5, d:0}", rec.Index.String())
				assert.EqualValues(t, "hello", rec.RawData)
			}*/

			// check second record
			if rec := <-res.RecordChan; assert.NotNil(t, rec) {
				assert.EqualValues(t, "{2.txt#200, len:5, d:-1}", rec.Index.String())
				assert.EqualValues(t, "hello", rec.RawData)
			}

			// check third record
			/*if rec := <-res.RecordChan; assert.NotNil(t, rec) {
				assert.EqualValues(t, "{3.txt#300, len:5, d:1}", rec.Index.String())
				assert.EqualValues(t, "hello", rec.RawData)
			}*/
		}
	}
}

// valid results + VIEW file (JSON array format)
func TestReaderJsonArrayView(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath := "/tmp/ryftprim-index-ja.txt"
	dataPath := "/tmp/ryfptrim-data-ja.bin"
	viewPath := "/tmp/ryfptrim-view-ja.bin"
	delimiter := "\r\n\f"

	//	defer os.RemoveAll(indexPath)
	//	defer os.RemoveAll(dataPath)
	//	defer os.RemoveAll(viewPath)

	rr := NewResultsReader(NewTask(nil, true), dataPath, indexPath, viewPath, delimiter)
	rr.CheckJsonArray = true
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true
	rr.MakeView = true

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(200 * time.Millisecond) // initial delay

		od, _ := os.Create(dataPath)
		oi, _ := os.Create(indexPath)
		assert.NotNil(t, od)
		assert.NotNil(t, oi)
		defer od.Close()
		defer oi.Close()

		testFakeRyftprim3j(od, oi, delimiter)

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
		assert.EqualValues(t, 3*(16+2+len(delimiter))+2, rr.totalDataLength)

		// check first record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{1.txt#100, len:16, d:0}", rec.Index.String())
			assert.EqualValues(t, `{"text":"hello"}`, rec.RawData)
		}

		// check second record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{2.txt#200, len:16, d:-1}", rec.Index.String())
			assert.EqualValues(t, `{"text":"hello"}`, rec.RawData)
		}

		// check third record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{3.txt#300, len:16, d:1}", rec.Index.String())
			assert.EqualValues(t, `{"text":"hello"}`, rec.RawData)
		}
	}

	// check the VIEW file created
	if vf, err := view.Open(viewPath); assert.NoError(t, err) {
		if assert.EqualValues(t, 3, vf.Count()) {
			// first record
			if iBeg, iEnd, dBeg, dEnd, err := vf.Get(0); assert.NoError(t, err) {
				assert.EqualValues(t, []int64{0, 24, 2, 18}, []int64{iBeg, iEnd, dBeg, dEnd})
			}

			// second record
			if iBeg, iEnd, dBeg, dEnd, err := vf.Get(1); assert.NoError(t, err) {
				assert.EqualValues(t, []int64{24, 41, 23, 39}, []int64{iBeg, iEnd, dBeg, dEnd})
			}

			// third record
			if iBeg, iEnd, dBeg, dEnd, err := vf.Get(2); assert.NoError(t, err) {
				assert.EqualValues(t, []int64{41, 65, 44, 60}, []int64{iBeg, iEnd, dBeg, dEnd})
			}

			// failed
			if iBeg, iEnd, dBeg, dEnd, err := vf.Get(3); assert.Error(t, err) {
				assert.EqualValues(t, []int64{-1, -1, -1, -1}, []int64{iBeg, iEnd, dBeg, dEnd})
				assert.Contains(t, err.Error(), "VIEW out of range")
			}
		}
	}

	// read with VIEW file
	if true {
		rr := NewResultsReader(NewTask(nil, true), dataPath, indexPath, viewPath, delimiter)
		rr.CheckJsonArray = true
		rr.RelativeToHome = "/ryftone"
		rr.OpenFilePollTimeout = 50 * time.Millisecond
		rr.ReadFilePollTimeout = 50 * time.Millisecond
		rr.ReadFilePollLimit = 20
		rr.ReadData = true
		rr.MakeView = false

		// emulate work:
		wg.Add(1)
		go func() {
			defer wg.Done()

			// soft stop
			time.Sleep(100 * time.Millisecond)
			rr.stop()
		}()

		res := search.NewResult()
		rr.process(res)
		wg.Wait()

		if assert.EqualValues(t, 0, res.ErrorsReported()) &&
			assert.EqualValues(t, 3, res.RecordsReported()) {

			// check first record
			if rec := <-res.RecordChan; assert.NotNil(t, rec) {
				assert.EqualValues(t, "{1.txt#100, len:16, d:0}", rec.Index.String())
				assert.EqualValues(t, `{"text":"hello"}`, rec.RawData)
			}

			// check second record
			if rec := <-res.RecordChan; assert.NotNil(t, rec) {
				assert.EqualValues(t, "{2.txt#200, len:16, d:-1}", rec.Index.String())
				assert.EqualValues(t, `{"text":"hello"}`, rec.RawData)
			}

			// check third record
			if rec := <-res.RecordChan; assert.NotNil(t, rec) {
				assert.EqualValues(t, "{3.txt#300, len:16, d:1}", rec.Index.String())
				assert.EqualValues(t, `{"text":"hello"}`, rec.RawData)
			}
		}
	}

	// read with VIEW file
	if true {
		rr := NewResultsReader(NewTask(nil, true), dataPath, indexPath, viewPath, delimiter)
		rr.RelativeToHome = "/ryftone"
		rr.CheckJsonArray = true
		rr.OpenFilePollTimeout = 50 * time.Millisecond
		rr.ReadFilePollTimeout = 50 * time.Millisecond
		rr.ReadFilePollLimit = 20
		rr.ReadData = true
		rr.MakeView = false
		rr.Offset = 1
		rr.Limit = 1

		// emulate work:
		wg.Add(1)
		go func() {
			defer wg.Done()

			// soft stop
			time.Sleep(100 * time.Millisecond)
			rr.stop()
		}()

		res := search.NewResult()
		rr.process(res)
		wg.Wait()

		if assert.EqualValues(t, 0, res.ErrorsReported()) &&
			assert.EqualValues(t, 1, res.RecordsReported()) {

			// check first record
			/*if rec := <-res.RecordChan; assert.NotNil(t, rec) {
				assert.EqualValues(t, "{1.txt#100, len:16, d:0}", rec.Index.String())
				assert.EqualValues(t, `{"text":"hello"}`, rec.RawData)
			}*/

			// check second record
			if rec := <-res.RecordChan; assert.NotNil(t, rec) {
				assert.EqualValues(t, "{2.txt#200, len:16, d:-1}", rec.Index.String())
				assert.EqualValues(t, `{"text":"hello"}`, rec.RawData)
			}

			// check third record
			/*if rec := <-res.RecordChan; assert.NotNil(t, rec) {
				assert.EqualValues(t, "{3.txt#300, len:16, d:1}", rec.Index.String())
				assert.EqualValues(t, `{"text":"hello"}`, rec.RawData)
			}*/
		}
	}
}

// valid results (no data read)
func TestReaderNoData(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
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
		time.Sleep(200 * time.Millisecond) // initial delay

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
			assert.EqualValues(t, "{1.txt#100, len:5, d:0}", rec.Index.String())
			assert.Nil(t, rec.RawData) // assert.EqualValues(t, "hello", rec.RawData)
		}

		// check second record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{2.txt#200, len:5, d:-1}", rec.Index.String())
			assert.Nil(t, rec.RawData) // assert.EqualValues(t, "hello", rec.RawData)
		}

		// check third record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{3.txt#300, len:5, d:1}", rec.Index.String())
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

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
	rr.RelativeToHome = "/ryftone"
	rr.OpenFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollTimeout = 50 * time.Millisecond
	rr.ReadFilePollLimit = 20
	rr.ReadData = true
	rr.Offset = 1 // skip 1 record
	rr.Limit = 2  // !!! only TWO records expected

	var wg sync.WaitGroup

	// emulate ryftprim work:
	// write fake INDEX/DATA files
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(200 * time.Millisecond) // initial delay

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
		assert.EqualValues(t, 3*(5+len(delimiter)), rr.totalDataLength)

		// check first record
		/*if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{1.txt#100, len:5, d:0}", rec.Index.String())
			assert.EqualValues(t, "hello", rec.RawData)
		}*/

		// check second record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{2.txt#200, len:5, d:-1}", rec.Index.String())
			assert.EqualValues(t, "hello", rec.RawData)
		}

		// check third record
		if rec := <-res.RecordChan; assert.NotNil(t, rec) {
			assert.EqualValues(t, "{3.txt#300, len:5, d:1}", rec.Index.String())
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

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
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
		time.Sleep(200 * time.Millisecond) // initial delay

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

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
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
		time.Sleep(200 * time.Millisecond) // initial delay

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

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
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
		time.Sleep(200 * time.Millisecond) // initial delay

		rr.cancel()
	}()

	res := search.NewResult()
	rr.process(res)
	wg.Wait()

	// log.Debugf("done, check results read")
	if assert.EqualValues(t, 0, res.ErrorsReported()) &&
		assert.EqualValues(t, 0, res.RecordsReported()) {
		assert.True(t, rr.isCancelled())
		assert.True(t, rr.isStopped())
	}
}

// bad results (failed to read INDEX)
func TestReaderFailedToReadIndex(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
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
		time.Sleep(200 * time.Millisecond) // initial delay

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

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
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
		time.Sleep(200 * time.Millisecond) // initial delay

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
		assert.True(t, rr.isCancelled())
		assert.True(t, rr.isStopped())
	}
}

// bad results (cancel to open DATA file)
func TestReaderCancelToOpenData(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
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
		assert.True(t, rr.isCancelled())
		assert.True(t, rr.isStopped())
	}
}

// bad results (failed to parse INDEX)
func TestReaderFailedToParseIndex(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
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
		time.Sleep(200 * time.Millisecond) // initial delay

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

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
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
		time.Sleep(200 * time.Millisecond) // initial delay

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

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
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
		time.Sleep(200 * time.Millisecond) // initial delay

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
		assert.True(t, rr.isCancelled())
		assert.True(t, rr.isStopped())
	}
}

// bad results (failed to read DATA delimiter)
func TestReaderFailedToReadDelim(t *testing.T) {
	SetLogLevelString(testLogLevel)

	indexPath, dataPath, delimiter := testReaderFake()
	defer os.RemoveAll(indexPath)
	defer os.RemoveAll(dataPath)

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
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
		time.Sleep(200 * time.Millisecond) // initial delay

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

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
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
		time.Sleep(200 * time.Millisecond) // initial delay

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

	rr := NewResultsReader(NewTask(nil, false), dataPath, indexPath, "", delimiter)
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
		time.Sleep(200 * time.Millisecond) // initial delay

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
		assert.True(t, rr.isCancelled())
		assert.True(t, rr.isStopped())
	}
}
