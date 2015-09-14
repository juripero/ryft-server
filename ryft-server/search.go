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

package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/getryft/ryft-rest-api/rol"
	"github.com/getryft/ryft-rest-api/ryft-server/encoder"
	"github.com/getryft/ryft-rest-api/ryft-server/transcoder"
	"github.com/getryft/ryft-rest-api/ryft-server/crpoll"
	"github.com/getryft/ryft-rest-api/ryft-server/names"
	"github.com/getryft/ryft-rest-api/ryft-server/records"
	"github.com/getryft/ryft-rest-api/ryft-server/srverr"
	"github.com/gin-gonic/gin"
)

func cleanup(file *os.File) {
	if file != nil {
		file.Close()
		if !KeepResults {
			os.Remove(file.Name())
		}
	}
}

type SearchParams struct {
	Query           string   `form:"query" binding:"required"`     // Search query, for example: ( RAW_TEXT CONTAINS "night" )
	Files           []string `form:"files" binding:"required"`     // Source files
	Surrounding     uint16   `form:"surrounding"`                  // Specifies the number of characters before the match and after the match that will be returned when the input specifier type is raw text
	Fuzziness       uint8    `form:"fuzziness"`                    // Is the fuzziness of the search. Measured as the maximum Hamming distance.
	Format          string   `form:"format"`                       // Source format parser name
	CaseSensitive   bool     `form:"cs"`                           // Case sensitive flag
}

func NewSearchParams() (p SearchParams) {
	p.Format = transcoder.RAWTRANSCODER
	return
}

func search(c *gin.Context) {
	defer srverr.DeferRecover(c)

	var err error

	// parse request parameters
	params := NewSearchParams()
	if err = c.Bind(&params); err != nil {
		panic(srverr.New(http.StatusBadRequest, err.Error()))
	}

	accept := c.NegotiateFormat(encoder.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = encoder.MIMEJSON
	}

	// setting up encoder to respond with requested format
	var enc encoder.Encoder
	if enc, err = encoder.GetByMimeType(accept); err != nil {
		panic(srverr.New(http.StatusBadRequest, err.Error()))
	}
	c.Header("Content-Type", accept)

	// setting up transcoder to convert raw data
	var tcode transcoder.Transcoder
	if tcode, err = transcoder.GetByFormat(params.Format); err != nil {
		panic(srverr.New(http.StatusBadRequest, err.Error()))
	}

	// get a new unique search index
	n := names.New()
	log.Printf("SEARCH(%d): %s", n.Index, c.Request.URL.String())

	p := progress(&params, n)

	// read an index file
	var idx, res *os.File
	if idx, err = crpoll.OpenFile(names.ResultsDirPath(n.IdxFile), p); err != nil {
		panic(srverr.New(http.StatusInternalServerError, err.Error()))
	}
	defer cleanup(idx)

	//read a results file
	if res, err = crpoll.OpenFile(names.ResultsDirPath(n.ResultFile), p); err != nil {
		panic(srverr.New(http.StatusInternalServerError, err.Error()))
	}
	defer cleanup(res)

	indexes, drop := records.Poll(idx, p)
	recs 	:= dataPoll(indexes, res)
	items, transcodeErrors := tcode.Transcode(recs)
	go logErrors("Transcode Error: %s", transcodeErrors)

	_ = drop

	streamAllRecords(c, enc, items)

}

func logErrors(format string,  errors chan error){
	for err := range errors {
		if err != nil {
			log.Printf(format, err.Error())
		}
	}
}

func streamAllRecords(c *gin.Context, enc encoder.Encoder, recs chan interface{}){
	first := true
	c.Stream(func(w io.Writer) bool {
		if first {
			enc.Begin(w)
			first = false
		}

		if record, ok := <-recs; ok {
//			log.Printf("RECORD: %+v", record)
			if err := enc.Write(w, record); err != nil {
				log.Panicln(err)
			} else {
				c.Writer.Flush()
			}
			return true
		} else {
			enc.End(w)
			return false
		}
	})
}


const (
	PollingInterval = time.Millisecond * 50
	PollBufferCapacity = 64
)

func dataPoll(input chan records.IdxRecord, dataFile *os.File) chan records.IdxRecord {
	output := make(chan records.IdxRecord, PollBufferCapacity)
	go func(){
		for rec := range input {
			rec.Data = nextData(dataFile, rec.Length)
			output <- rec
		}
		close(output)
	}()
	return output
}

func nextData(res *os.File, length uint16) (result []byte) {
	var total uint16 = 0
	for total < length {
		data := make([]byte, length-total)
		n, _ := res.Read(data)
		if n != 0 {
			result = append(result, data...)
			total = total + uint16(n)
		} else {
			time.Sleep(PollingInterval)
		}
	}
	return
}


func progress(s *SearchParams, n names.Names) (ch chan error) {
	ch = make(chan error, 1)
	go func() {
		ds := rol.RolDSCreate()
		defer ds.Delete()

		for _, f := range s.Files {
			ok := ds.AddFile(f)
			if !ok {
				ch <- srverr.New(http.StatusNotFound, "Could not add file "+f)
				return
			}
		}

		idxFile := names.PathInRyftoneForResultDir(n.IdxFile)

		resultsDs := ds.SearchFuzzyHamming(names.PathInRyftoneForResultDir(n.ResultFile), s.Query, s.Surrounding, s.Fuzziness, "", &idxFile, s.CaseSensitive)
		log.Printf("PROGRESS(%d): COMPLETE.", n.Index)

		defer resultsDs.Delete()

		if err := resultsDs.HasErrorOccured(); err != nil {
			if !err.IsStrangeError() {
				ch <- srverr.New(http.StatusInternalServerError, err.Error())
				return
			}
		}

		ch <- nil

	}()
	return
}
