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

package rest

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/getryft/ryft-server/rest/codec"
	my_codec "github.com/getryft/ryft-server/rest/codec/msgpack.v1"
	"github.com/getryft/ryft-server/rest/format"
	my_format "github.com/getryft/ryft-server/rest/format/raw"
	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/view"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// SearchShowParams contains all the bound parameters for the /search/show endpoint.
type SearchShowParams struct {
	DataFile  string `form:"data" json:"data,omitempty" msgpack:"data,omitempty"`
	IndexFile string `form:"index" json:"index,omitempty" msgpack:"index,omitempty"`
	ViewFile  string `form:"view" json:"view,omitempty" msgpack:"view,omitempty"`
	Delimiter string `form:"delimiter" json:"delimiter,omitempty" msgpack:"delimiter,omitempty"`
	Session   string `form:"session" json:"session,omitempty" msgpack:"session,omitempty"`
	Offset    uint64 `form:"offset" json:"offset,omitempty" msgpack:"offset,omitempty"`
	Count     uint64 `form:"count" json:"count,omitempty" msgpack:"count,omitempty"`

	Format string `form:"format" json:"format,omitempty" msgpack:"format,omitempty"`
	Fields string `form:"fields" json:"fields,omitempty" msgpack:"fields,omitempty"` // for XML and JSON formats
	Stream bool   `form:"stream" json:"stream,omitempty" msgpack:"stream,omitempty"`

	Local bool `form:"local" json:"local,omitempty" msgpack:"local,omitempty"`

	// internal parameters
	InternalErrorPrefix bool `form:"--internal-error-prefix" json:"-" msgpack:"-"` // include host prefixes for error messages
	//InternalNoSessionId bool `form:"--internal-no-session-id"`

	// private configuration
	relativeToHome string
	updateHostTo   string
}

// Handle /search/show endpoint.
func (server *Server) DoSearchShow(ctx *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(ctx)

	var err error

	// parse request parameters
	params := SearchShowParams{
		Format: format.RAW,
	}
	if err := binding.Form.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	var sessionInfo []interface{}
	if len(params.Session) != 0 {
		session, err := ParseSession(server.Config.Sessions.Secret, params.Session)
		if err != nil {
			panic(NewError(http.StatusBadRequest, err.Error()).
				WithDetails("failed to parse session token"))
		}

		if info, ok := session.GetData("info").([]interface{}); !ok {
			panic(NewError(http.StatusBadRequest, "invalid data format").
				WithDetails("failed to parse session token"))
		} else {
			sessionInfo = info
		}
	}

	params.Delimiter = mustParseDelim(params.Delimiter)
	if format.IsNull(params.Format) {
		params.DataFile = "" // no DATA for NULL format
	}

	// setting up transcoder to convert raw data
	// XML and JSON support additional fields filtration
	var tcode format.Format
	tcode_opts := map[string]interface{}{
		"fields": params.Fields,
	}
	if tcode, err = format.New(params.Format, tcode_opts); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to get transcoder"))
	}

	accept := ctx.NegotiateFormat(codec.GetSupportedMimeTypes()...)
	if accept == "" { // default to JSON
		accept = codec.MIME_JSON
		// log.Debugf("[%s]: Content-Type changed to %s", CORE, accept)
	}
	ctx.Header("Content-Type", accept)

	// setting up encoder to respond with requested format
	// we can use two formats:
	// - single JSON value (not appropriate for large data set)
	// - with tags to report data records and the statistics in a stream
	enc, err := codec.NewEncoder(ctx.Writer, accept, params.Stream)
	if err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to get encoder"))
	}
	ctx.Set("encoder", enc) // to recover from panic in appropriate format

	// get search engine
	userName, authToken, homeDir, userTag := server.parseAuthAndHome(ctx)
	mountPoint, err := server.getMountPoint(homeDir)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get mount point"))
	}
	mountPoint = filepath.Join(mountPoint, homeDir)

	var res *search.Result
	log.WithFields(map[string]interface{}{
		"user":    userName,
		"home":    homeDir,
		"cluster": userTag,
	}).Infof("[%s]: start GET /search/show", CORE)
	defer log.WithField("result", res).Infof("[%s]: /search/show done", CORE)

	params.relativeToHome = mountPoint
	params.updateHostTo = server.Config.HostName

	if params.Local || len(sessionInfo) <= 1 {
		var nodes []nodeSearchShow
		nodes, err = server.searchShowGetNodes(sessionInfo, params)
		if err != nil {
			panic(NewError(http.StatusInternalServerError, err.Error()).
				WithDetails("failed to get cluster nodes"))
		}
		for _, node := range nodes {
			if node.isLocal {
				res, err = doLocalSearchShow(mountPoint, node.params)
				break
			}
		}
	} else {
		var nodes []nodeSearchShow
		nodes, err = server.searchShowGetNodes(sessionInfo, params)
		if err != nil {
			panic(NewError(http.StatusInternalServerError, err.Error()).
				WithDetails("failed to get cluster nodes"))
		}
		res, err = doRemoteSearchShowMux(mountPoint, authToken, nodes)
	}
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get search results"))
	}
	if res == nil {
		panic(NewError(http.StatusInternalServerError, "no results available").
			WithDetails("failed to get search results"))
	}

	// in case of unexpected panic
	// we need to cancel search request
	// to prevent resource leaks
	defer cancelIfNotDone(res)

	// ctx.Stream() logic
	var lastError error

	// error prefix
	var errorPrefix string
	if params.InternalErrorPrefix {
		errorPrefix = server.Config.HostName
	}

	// put error to stream
	putErr := func(err_ error) {
		// to distinguish nodes in cluster mode
		// mark all errors with a prefix
		if len(errorPrefix) != 0 {
			err_ = fmt.Errorf("[%s]: %s", errorPrefix, err_)
		}
		err := enc.EncodeError(err_)
		if err != nil {
			panic(err)
		}
		lastError = err_
	}

	// put record to stream
	putRec := func(rec *search.Record) {
		xrec := tcode.FromRecord(rec)
		if xrec != nil {
			err = enc.EncodeRecord(xrec)
			if err != nil {
				panic(err)
			}
			// ctx.Writer.Flush() // TODO: check performance!!!
		}
	}

	// process results!
	for {
		select {
		case <-ctx.Writer.CloseNotify(): // cancel processing
			log.Warnf("[%s]: cancelling by user (connection is gone)...", CORE)
			if errors, records := res.Cancel(); errors > 0 || records > 0 {
				log.WithFields(map[string]interface{}{
					"errors":  errors,
					"records": records,
				}).Debugf("[%s]: some errors/records are ignored", CORE)
			}
			return // cancelled

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				putRec(rec)
			}

		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				putErr(err)
			}

		case <-res.DoneChan:
			// drain the records...
			for rec := range res.RecordChan {
				putRec(rec)
			}

			// ... and errors
			for err := range res.ErrorChan {
				putErr(err)
			}

			// special case: if no records and no stats were received
			// but just an error, we panic to return 500 status code
			if res.RecordsReported() == 0 && res.Stat == nil &&
				res.ErrorsReported() == 1 && lastError != nil {
				panic(lastError)
			}

			// close encoder
			err := enc.Close()
			if err != nil {
				panic(err)
			}

			return // done
		}
	}
}

// do local show
func doLocalSearchShow(mountPoint string, params SearchShowParams) (*search.Result, error) {
	if len(params.ViewFile) != 0 {
		return doLocalSearchShowView(mountPoint, params)
	}

	return doLocalSearchShowNoView(mountPoint, params)
}

// do local show (no view file)
func doLocalSearchShowNoView(mountPoint string, params SearchShowParams) (*search.Result, error) {
	var idxFd, datFd *os.File
	var idxRd, datRd *bufio.Reader
	var dataPos uint64 // DATA read position

	// INDEX file reader
	if idxRd == nil {
		// try to open INDEX file
		f, err := os.Open(filepath.Join(mountPoint, params.IndexFile))
		if err != nil {
			return nil, fmt.Errorf("failed to open INDEX file: %s", err)
		}

		idxFd, idxRd = f, bufio.NewReaderSize(f, 256*1024)
		// log.Debugf("[%s/show]: open INDEX at: %s", CORE, f.Name())
	}

	// DATA file reader
	if datRd == nil && len(params.DataFile) != 0 {
		// try to open DATA file
		f, err := os.Open(filepath.Join(mountPoint, params.DataFile))
		if err != nil {
			return nil, fmt.Errorf("failed to open DATA file: %s", err)
		}

		datFd, datRd = f, bufio.NewReaderSize(f, 256*1024)
		// log.Debugf("[%s/show]: open DATA at: %s", CORE, f.Name())
	}

	res := search.NewResult()
	go func() {
		defer res.Close()
		defer res.ReportDone()

		// close at the end
		if idxFd != nil {
			defer idxFd.Close()
		}
		if datFd != nil {
			defer datFd.Close()
		}

		// buffer to check delimiter
		delim := make([]byte, len(params.Delimiter))

		var i uint64
		for i = 0; !res.IsCancelled(); i++ {
			// read line by line
			line, err := idxRd.ReadBytes('\n')
			if err != nil {
				if err == io.EOF && 0 == len(line) {
					return // DONE
				} else {
					res.ReportError(fmt.Errorf("failed to read INDEX: %s", err))
					return // FAILED
				}
			}

			// parse index
			index, err := search.ParseIndex(line)
			if err != nil {
				res.ReportError(fmt.Errorf("failed to parse INDEX: %s", err))
				return // FAILED
			}

			//log.Debugf("[%s/show]: read INDEX: %s", CORE, index)

			// skip requested number of records
			if i < params.Offset {
				if datRd != nil {
					n := int(index.Length) + len(params.Delimiter)
					m, err := datRd.Discard(n)
					if err != nil {
						res.ReportError(fmt.Errorf("failed to skip DATA: %s", err))
						return // FAILED
					} else if m != n {
						res.ReportError(fmt.Errorf("not all DATA skipped: %d of %d", m, n))
						return // FAILED
					}
					dataPos += uint64(m)
				}
				continue // go to next RECORD
			}

			var data []byte
			if datRd != nil {
				data = make([]byte, int(index.Length))
				m, err := io.ReadFull(datRd, data)
				if err != nil {
					res.ReportError(fmt.Errorf("failed to read DATA: %s", err))
					return // FAILED
				} else if m != len(data) {
					res.ReportError(fmt.Errorf("not all DATA read: %d of %d", m, len(data)))
					return // FAILED
				}
				dataPos += index.Length

				// log.Debugf("[%s/show]: DATA: %s of %d bytes", CORE, data, index.Length)

				// read and check delimiter
				if len(params.Delimiter) > 0 {
					// or just ... datRd.Discard(len(rr.Delimiter))

					// try to read delimiter
					m, err := io.ReadFull(datRd, delim)
					if err != nil {
						res.ReportError(fmt.Errorf("failed to read DATA delimiter: %s", err))
						return // FAILED
					} else if m != len(delim) {
						res.ReportError(fmt.Errorf("not all DATA delimiter read: %d of %d", m, len(delim)))
						return // FAILED
					}

					// log.Debugf("[%s/show]: DATA delim: %x of %d bytes", CORE, delim, m)

					// check delimiter expected
					if string(delim) != params.Delimiter {
						res.ReportError(fmt.Errorf("%q unexpected delimiter found at %d", string(delim), dataPos))
						return // FAILED
					}

					dataPos += uint64(len(delim))
				}
			} // dataRd

			// trim mount point from file name!
			if len(params.relativeToHome) != 0 {
				if rel, err := filepath.Rel(params.relativeToHome, index.File); err == nil {
					index.File = rel
				} else {
					// keep the absolute filepath as fallback
					// log.WithError(err).Debugf("[%s/show]: failed to get relative path", TAG)
				}
			}

			// update host for cluster mode!
			index.UpdateHost(params.updateHostTo)

			// report new record
			rec := search.NewRecord(index, data)
			// log.WithField("rec", rec).Debugf("[%s/show]: new record", TAG) // FIXME: DEBUG

			res.ReportRecord(rec)
			if params.Count > 0 && res.RecordsReported() >= params.Count {
				// log.WithField("limit", params.Count).Debugf("[%s/show]: stopped by limit", TAG)
				return // DONE
			}
		}
	}()

	return res, nil // OK for now
}

// do local show (view file)
func doLocalSearchShowView(mountPoint string, params SearchShowParams) (*search.Result, error) {
	var idxFd, datFd *os.File
	var idxRd, datRd *bufio.Reader
	var indexPos int64 // INDEX read position
	var dataPos int64  // DATA read position

	// INDEX file reader
	if idxRd == nil {
		// try to open INDEX file
		f, err := os.Open(filepath.Join(mountPoint, params.IndexFile))
		if err != nil {
			return nil, fmt.Errorf("failed to open INDEX file: %s", err)
		}

		idxFd, idxRd = f, bufio.NewReaderSize(f, 256*1024)
		// log.Debugf("[%s/show]: open INDEX at: %s", CORE, f.Name())
	}

	// DATA file reader
	if datRd == nil && len(params.DataFile) != 0 {
		// try to open DATA file
		f, err := os.Open(filepath.Join(mountPoint, params.DataFile))
		if err != nil {
			return nil, fmt.Errorf("failed to open DATA file: %s", err)
		}

		datFd, datRd = f, bufio.NewReaderSize(f, 256*1024)
		// log.Debugf("[%s/show]: open DATA at: %s", CORE, f.Name())
	}

	// VIEW file reader
	var viewRd *view.Reader
	if viewRd == nil && len(params.ViewFile) != 0 {
		f, err := view.Open(filepath.Join(mountPoint, params.ViewFile))
		if err != nil {
			return nil, fmt.Errorf("failed to open VIEW file: %s", err)
		}

		viewRd = f
	}

	res := search.NewResult()
	go func() {
		defer res.Close()
		defer res.ReportDone()

		// close at the end
		if idxFd != nil {
			defer idxFd.Close()
		}
		if datFd != nil {
			defer datFd.Close()
		}
		if viewRd != nil {
			defer viewRd.Close()
		}

		// buffer to check delimiter
		delim := make([]byte, len(params.Delimiter))

		// adjust count: if zero - get rest of records
		if n := viewRd.Count(); true {
			if params.Count == 0 {
				params.Count = n
			}

			// limit count to available number of items
			if params.Offset+params.Count > n {
				params.Count = n - params.Offset
			}
		}

		var i uint64
		for i = 0; i < params.Count && !res.IsCancelled(); i++ {
			indexBeg, indexEnd, dataBeg, dataEnd, err := viewRd.Get(int64(i + params.Offset))
			if err != nil {
				res.ReportError(fmt.Errorf("failed to read VIEW: %s", err))
				return // FAILED
			}

			// read INDEX line
			if n := int(indexBeg - indexPos); n >= 0 && n < idxRd.Buffered() {
				if n != 0 {
					// we are within one buffer range, so just discard
					//fmt.Printf("discarding %d bytes", n)
					if _, err := idxRd.Discard(n); err != nil {
						res.ReportError(fmt.Errorf("failed to seek INDEX file: %s", err))
						return // FAILED
					}
				}
			} else {
				// base case. read before buffer or too far after...
				//fmt.Printf("seek to %d bytes (rpos: %d)", fpos, r.rpos)
				if _, err := idxFd.Seek(indexBeg, io.SeekStart); err != nil {
					res.ReportError(fmt.Errorf("failed to seek INDEX file: %s", err))
					return // FAILED
				}

				// have to reset buffer
				idxRd.Reset(idxFd)
				indexPos = indexBeg
			}

			line := make([]byte, int(indexEnd-indexBeg))
			_, err = io.ReadFull(idxRd, line)
			if err != nil {
				if err == io.EOF && 0 == len(line) {
					return // DONE
				} else {
					res.ReportError(fmt.Errorf("failed to read INDEX(view): %s", err))
					return // FAILED
				}
			}
			indexPos += int64(len(line))

			// parse index
			index, err := search.ParseIndex(line)
			if err != nil {
				res.ReportError(fmt.Errorf("failed to parse INDEX: %s", err))
				return // FAILED
			}

			//log.Debugf("[%s/show]: read INDEX: %s", CORE, index)

			var data []byte
			if datRd != nil {
				if index.Length != uint64(dataEnd-dataBeg) {
					res.ReportError(fmt.Errorf("INDEX and VIEW mismatch: %d != %d (expected)", dataEnd-dataBeg, index.Length))
					return // FAILED
				}

				if n := int(dataBeg - dataPos); n >= 0 && n < datRd.Buffered() {
					if n != 0 {
						// we are within one buffer range, so just discard
						//fmt.Printf("DATA: discarding %d bytes", n)
						if _, err := datRd.Discard(n); err != nil {
							res.ReportError(fmt.Errorf("failed to seek DATA file: %s", err))
							return // FAILED
						}
					}
				} else {
					// base case. read before buffer or too far after...
					//fmt.Printf("seek to %d bytes (rpos: %d)", fpos, r.rpos)
					if _, err := datFd.Seek(dataBeg, io.SeekStart); err != nil {
						res.ReportError(fmt.Errorf("failed to seek DATA file: %s", err))
						return // FAILED
					}

					// have to reset buffer
					datRd.Reset(datFd)
					dataPos = dataBeg
				}

				data = make([]byte, int(index.Length))
				m, err := io.ReadFull(datRd, data)
				if err != nil {
					res.ReportError(fmt.Errorf("failed to read DATA: %s", err))
					return // FAILED
				} else if m != len(data) {
					res.ReportError(fmt.Errorf("not all DATA read: %d of %d", m, len(data)))
					return // FAILED
				}
				dataPos += int64(index.Length)

				// log.Debugf("[%s/show]: DATA: %s of %d bytes", CORE, data, index.Length)

				// read and check delimiter
				if len(params.Delimiter) > 0 {
					// or just ... datRd.Discard(len(rr.Delimiter))

					// try to read delimiter
					m, err := io.ReadFull(datRd, delim)
					if err != nil {
						res.ReportError(fmt.Errorf("failed to read DATA delimiter: %s", err))
						return // FAILED
					} else if m != len(delim) {
						res.ReportError(fmt.Errorf("not all DATA delimiter read: %d of %d", m, len(delim)))
						return // FAILED
					}

					// log.Debugf("[%s/show]: DATA delim: %x of %d bytes", CORE, delim, m)

					// check delimiter expected
					if string(delim) != params.Delimiter {
						res.ReportError(fmt.Errorf("%q unexpected delimiter found at %d", string(delim), dataPos))
						return // FAILED
					}

					dataPos += int64(len(delim))
				}
			} // dataRd

			// trim mount point from file name!
			if len(params.relativeToHome) != 0 {
				if rel, err := filepath.Rel(params.relativeToHome, index.File); err == nil {
					index.File = rel
				} else {
					// keep the absolute filepath as fallback
					// log.WithError(err).Debugf("[%s/show]: failed to get relative path", TAG)
				}
			}

			// update host for cluster mode!
			index.UpdateHost(params.updateHostTo)

			// report new record
			rec := search.NewRecord(index, data)
			// log.WithField("rec", rec).Debugf("[%s/show]: new record", TAG) // FIXME: DEBUG

			res.ReportRecord(rec)
			if params.Count > 0 && res.RecordsReported() >= params.Count {
				// log.WithField("limit", params.Count).Debugf("[%s/show]: stopped by limit", TAG)
				return // DONE
			}
		}
	}()

	return res, nil // OK for now
}

// do remote show via HTTP request
func doRemoteSearchShowHttp(serverUrl, authToken string, params SearchShowParams) (*search.Result, error) {
	// server URL should be parsed in engine initialization
	// so we can omit error checking here
	u, _ := url.Parse(serverUrl)
	u.Path += "/search/show"

	// prepare query
	q := url.Values{}
	if !format.IsNull(params.Format) {
		q.Set("format", "raw")
	} else {
		q.Set("format", "null")
	}
	q.Set("local", fmt.Sprintf("%t", true))
	q.Set("stream", fmt.Sprintf("%t", true))
	q.Set("--internal-error-prefix", fmt.Sprintf("%t", true)) // enable error prefixes!

	q.Set("data", params.DataFile)
	q.Set("index", params.IndexFile)
	q.Set("view", params.ViewFile)
	if len(params.Delimiter) != 0 {
		q.Set("delimiter", params.Delimiter)
	}
	q.Set("offset", fmt.Sprintf("%d", params.Offset))
	q.Set("count", fmt.Sprintf("%d", params.Count))
	u.RawQuery = q.Encode()

	// prepare request
	// task.log().WithField("url", u.String()).Infof("[%s]: sending GET", TAG)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		// task.log().WithError(err).Warnf("[%s]: failed to create request", TAG)
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	// we expect MSGPACK format for streaming
	req.Header.Set("Accept", my_codec.MIME)

	// authorization
	if len(authToken) != 0 {
		req.Header.Set("Authorization", authToken)
	}

	res := search.NewResult()
	go func() {
		// some futher cleanup
		defer res.Close()
		defer res.ReportDone()

		doneCh := make(chan struct{})
		defer close(doneCh)

		cancelCh := make(chan struct{})
		req.Cancel = cancelCh
		var cancelled int32 // atomic

		// do HTTP request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			//task.log().WithError(err).Warnf("[%s]: failed to send request", TAG)
			res.ReportError(fmt.Errorf("failed to send request: %s", err))
			return // failed
		}

		defer resp.Body.Close() // close it later

		// check status code
		if resp.StatusCode != http.StatusOK {
			// task.log().WithField("status", resp.StatusCode).Warnf("[%s]: invalid response status", TAG)
			res.ReportError(fmt.Errorf("invalid response status: %d (%s)", resp.StatusCode, resp.Status))
			return // failed (not 200)
		}

		// read response and report records and/or statistics
		dec, _ := my_codec.NewStreamDecoder(resp.Body)

		// handle task cancellation
		go func() {
			select {
			case <-res.CancelChan:
				// task.log().Warnf("[%s]: cancelling by client", TAG)
				if atomic.CompareAndSwapInt32(&cancelled, 0, 1) {
					close(cancelCh) // cancel the request, once
				}

			case <-doneCh:
				// task.log().Debugf("[%s]: done", TAG)
				return
			}
		}()

		// read stream of tag-object pairs
		for atomic.LoadInt32(&cancelled) == 0 {
			tag, err := dec.NextTag()
			if err != nil {
				//task.log().WithError(err).Warnf("[%s]: failed to decode next tag", TAG)
				res.ReportError(fmt.Errorf("failed to decode next tag: %s", err))
				return // failed
			}

			switch tag {
			case my_codec.TAG_EOF:
				//task.log().WithField("result", res).Infof("[%s]: got end of response", TAG)
				return // DONE

			case my_codec.TAG_REC:
				item := my_format.NewRecord()
				if err := dec.Next(item); err != nil {
					//task.log().WithError(err).Warnf("[%s]: failed to decode record", TAG)
					res.ReportError(fmt.Errorf("failed to decode record: %s", err))
					return // failed
				} else {
					rec := my_format.ToRecord(item)
					// task.log().WithField("rec", rec).Debugf("[%s]: new record received", TAG) // FIXME: DEBUG
					res.ReportRecord(rec)
					// continue
				}

			case my_codec.TAG_ERR:
				var msg string
				if err := dec.Next(&msg); err != nil {
					//task.log().WithError(err).Warnf("[%s]: failed to decode error", TAG)
					res.ReportError(fmt.Errorf("failed to decode error: %s", err))
					return // failed
				} else {
					err := fmt.Errorf("%s", msg)
					// task.log().WithError(err).Debugf("[%s]: new error received", TAG) // FIXME: DEBUG
					res.ReportError(err)
					// continue
				}

			case my_codec.TAG_STAT:
				stat := my_format.NewStat()
				if err := dec.Next(stat); err != nil {
					//task.log().WithError(err).Warnf("[%s]: failed to decode statistics", TAG)
					res.ReportError(fmt.Errorf("failed to decode statistics: %s", err))
					return // failed
				} else {
					res.Stat = my_format.ToStat(stat)
					// task.log().WithField("stat", res.Stat).Debugf("[%s]: statistics received", TAG) // FIXME: DEBUG
					// continue
				}

			default:
				// task.log().WithField("tag", tag).Warnf("[%s]: unknown tag", TAG)
				res.ReportError(fmt.Errorf("unknown data tag received: %v", tag))
				return // failed, no sense to continue processing
			}
		}
	}()

	return res, nil // OK for now
}

type nodeSearchShow struct {
	isLocal bool
	nodeUrl string
	params  SearchShowParams
}

// get nodes according to incoming info and offset/count
func (s *Server) searchShowGetNodes(info []interface{}, params SearchShowParams) ([]nodeSearchShow, error) {
	params.Session = ""
	res := make([]nodeSearchShow, 0, len(info))
	if len(info) <= 1 {
		nss := nodeSearchShow{
			isLocal: true,
			nodeUrl: "", // not used
			params:  params,
		}

		if len(info) != 0 {
			node_ := info[0]
			if node, ok := node_.(map[string]interface{}); ok {
				if len(nss.params.DataFile) == 0 {
					nss.params.DataFile, _ = utils.AsString(node["data"])
				}
				if len(nss.params.IndexFile) == 0 {
					nss.params.IndexFile, _ = utils.AsString(node["index"])
				}
				if len(nss.params.ViewFile) == 0 {
					nss.params.ViewFile, _ = utils.AsString(node["view"])
				}
				if len(nss.params.Delimiter) == 0 {
					nss.params.Delimiter, _ = utils.AsString(node["delim"])
				}
			}
		}

		res = append(res, nss)
	} else {
		log.Debugf("requested range [%d..%d)", params.Offset, params.Offset+params.Count)
		// TODO: case if params.Count = 0 - means from offset till the END
		var offset uint64
		for _, node_ := range info {
			if node, ok := node_.(map[string]interface{}); ok {
				matches, _ := utils.AsUint64(node["matches"])
				beg := offset
				end := beg + matches
				offset += matches

				log.Debugf("  node: %v", node)

				if params.Offset+params.Count < beg || end <= params.Offset {
					continue // out of range
				}

				nss := nodeSearchShow{
					params: params,
				}
				if location, err := utils.AsString(node["location"]); err != nil {
					return nil, fmt.Errorf("failed to get location: %s", err)
				} else if len(location) != 0 {
					u, err := url.Parse(location)
					if err != nil {
						return nil, fmt.Errorf("failed to parse location: %s", err)
					}
					if s.isLocalServiceUrl(u) {
						nss.isLocal = true
						nss.nodeUrl = ""
					} else {
						nss.isLocal = false
						nss.nodeUrl = location
					}
				} else {
					nss.isLocal = true
					nss.nodeUrl = ""
				}

				nss.params.DataFile, _ = utils.AsString(node["data"])
				nss.params.IndexFile, _ = utils.AsString(node["index"])
				nss.params.ViewFile, _ = utils.AsString(node["view"])
				nss.params.Delimiter, _ = utils.AsString(node["delim"])
				if beg < params.Offset {
					nss.params.Offset = params.Offset - beg
				} else {
					nss.params.Offset = 0
				}
				if end < params.Offset+params.Count {
					nss.params.Count = matches - nss.params.Offset
				} else {
					nss.params.Count = matches - (end - (params.Offset + params.Count)) - nss.params.Offset
				}

				log.Debugf("%s mapped range [%d/%d]", nss.nodeUrl, nss.params.Offset, nss.params.Count)

				res = append(res, nss)
			} else {
				return nil, fmt.Errorf("bad info data format: %T", node_)
			}
		}
	}

	return res, nil // OK
}

// process and wait all
func doRemoteSearchShowMux(mountPoint string, authToken string, nodes []nodeSearchShow) (*search.Result, error) {
	mux := search.NewResult()

	var subtasks sync.WaitGroup
	results := make([]*search.Result, 0, len(nodes))

	// prepare requests
	for _, node := range nodes {
		var res *search.Result
		var err error
		if node.isLocal {
			res, err = doLocalSearchShow(mountPoint, node.params)
		} else {
			res, err = doRemoteSearchShowHttp(node.nodeUrl, authToken, node.params)
		}
		if err != nil {
			//task.log().WithError(err).Warnf("[%s]: failed to start /search backend", TAG)
			mux.ReportError(fmt.Errorf("failed to start /search backend: %s", err))
			continue
		}

		subtasks.Add(1)
		results = append(results, res)
	}

	go func() {
		// some futher cleanup
		defer mux.Close()
		defer mux.ReportDone()

		// communication channel to report completed results
		resCh := make(chan *search.Result, len(results))

		// start multiplexing results and errors
		//task.log().Debugf("[%s]: start subtask processing...", TAG)
		//var recordsReported uint64 // for all subtasks, atomic
		for _, res := range results {
			go func(res *search.Result) {
				defer func() {
					subtasks.Done()
					resCh <- res
				}()

				// drain subtask's records and errors
				for {
					select {
					case err, ok := <-res.ErrorChan:
						if ok && err != nil {
							// TODO: mark error with subtask's tag?
							// task.log().WithError(err).Debugf("[%s]: new error received", TAG) // FIXME: DEBUG
							mux.ReportError(err)
						}

					case rec, ok := <-res.RecordChan:
						if ok && rec != nil {
							if true { //atomic.AddUint64(&recordsReported, 1) <= recordsLimit {
								// task.log().WithField("rec", rec).Debugf("[%s]: new record received", TAG) // FIXME: DEBUG
								//rec.Index.UpdateHost(engine.IndexHost) // cluster mode!
								mux.ReportRecord(rec)
								/*} else {
								task.log().WithField("limit", recordsLimit).Infof("[%s]: stopped by limit", TAG)
								errors, records := res.Cancel()
								if errors > 0 || records > 0 {
									task.log().WithFields(map[string]interface{}{
										"errors":  errors,
										"records": records,
									}).Debugf("[%s]: some errors/records are ignored", TAG)
								}
								return // done!*/
							}
						}

					case <-res.DoneChan:
						// drain the whole errors channel
						for err := range res.ErrorChan {
							// task.log().WithError(err).Debugf("[%s]: *** new error received", TAG) // FIXME: DEBUG
							mux.ReportError(err)
						}

						// drain the whole records channel
						for rec := range res.RecordChan {
							if true { // atomic.AddUint64(&recordsReported, 1) <= recordsLimit {
								// task.log().WithField("rec", rec).Debugf("[%s]: *** new record received", TAG) // FIXME: DEBUG
								// rec.Index.UpdateHost(engine.IndexHost) // cluster mode!
								mux.ReportRecord(rec)
								/*} else {
								task.log().WithField("limit", recordsLimit).Infof("[%s]: *** stopped by limit", TAG)
								errors, records := res.Cancel()
								if errors > 0 || records > 0 {
									task.log().WithFields(map[string]interface{}{
										"errors":  errors,
										"records": records,
									}).Debugf("[%s]: *** some errors/records are ignored", TAG)
								}
								return // done!*/
							}
						}

						return // done!
					}
				}

			}(res)
		}

		// wait for statistics and process cancellation
		finished := make(map[*search.Result]bool)
	WaitLoop:
		for _ = range results {
			select {
			case res, ok := <-resCh:
				if ok && res != nil {
					// once subtask is finished combine statistics
					//					task.log().WithField("result", res).
					//						Infof("[%s]: subtask is finished", TAG)
					if res.Stat != nil {
						if mux.Stat == nil {
							// create multiplexed statistics
							mux.Stat = search.NewStat("")
						}
						mux.Stat.Merge(res.Stat)
					}
					finished[res] = true
				}
				continue WaitLoop

			case <-mux.CancelChan:
				// cancel all unfinished tasks
				// task.log().Warnf("[%s]: cancelling by client", TAG)
				for _, r := range results {
					if !finished[r] {
						errors, records := r.Cancel()
						if errors > 0 || records > 0 {
							log.WithFields(map[string]interface{}{
								"errors":  errors,
								"records": records,
							}).Debugf("[%s]: subtask is cancelled, some errors/records are ignored", CORE)
						}
					}
				}
				break WaitLoop
			}
		}

		// wait all goroutines
		log.Debugf("[%s]: waiting all subtasks...", CORE)
		subtasks.Wait()
		log.Debugf("[%s]: done", CORE)
	}()

	return mux, nil // OK for now
}
