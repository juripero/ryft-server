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
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getryft/ryft-server/crpoll"
	"github.com/getryft/ryft-server/encoder"
	"github.com/getryft/ryft-server/names"
	"github.com/getryft/ryft-server/records"
	//	"github.com/getryft/ryft-server/rol"
	"github.com/getryft/ryft-server/srverr"
	"github.com/getryft/ryft-server/transcoder"
	"github.com/gin-gonic/gin"
)

func cleanup(file *os.File) {
	if file != nil {
		log.Printf(" Close file %v", file.Name())
		file.Close()
		if !*KeepResults {
			os.Remove(file.Name())
		}
	}
}

const sepSign string = ","

/*
SearchParams contains all the bound params for the search operation
*/
type SearchParams struct {
	Query         string   `form:"query" json:"query" binding:"required"`
	Files         []string `form:"files" json:"files" binding:"required"`
	Surrounding   uint16   `form:"surrounding" json:"surrounding"`
	Fuzziness     uint8    `form:"fuzziness" json:"fuzziness"`
	Format        string   `form:"format" json:"format"`
	CaseSensitive bool     `form:"cs" json:"cs"`
	Fields        string   `form:"fields" json:"fields"`
	Local         bool     `form:"local" json:"local"`
	Nodes         uint8    `form:"nodes" json:"nodes"`
}

func search(c *gin.Context) {

	defer srverr.DeferRecover(c)

	var err error

	// parse request parameters
	params := SearchParams{}
	params.Format = transcoder.RAWTRANSCODER
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

	ryftParams := &RyftprimParams{
		Query:         params.Query,
		Files:         params.Files,
		Surrounding:   params.Surrounding,
		Fuzziness:     params.Fuzziness,
		Format:        params.Format,
		CaseSensitive: params.CaseSensitive,
		Fields:        params.Fields,
		Nodes:         params.Nodes,
	}
	p, statistic := ryftprim(ryftParams, &n)
	m := <-statistic

	// read an index file
	var idx, res *os.File
	if idx, err = crpoll.OpenFile(names.ResultsDirPath(n.IdxFile), p); err != nil {
		if serr, ok := err.(*srverr.ServerError); ok {
			panic(serr)
		} else {
			panic(srverr.New(http.StatusInternalServerError, err.Error()))
		}
	}
	defer cleanup(idx)

	//read a results file
	if res, err = crpoll.OpenFile(names.ResultsDirPath(n.ResultFile), p); err != nil {
		if serr, ok := err.(*srverr.ServerError); ok {
			panic(serr)
		} else {
			panic(srverr.New(http.StatusInternalServerError, err.Error()))
		}
	}
	defer cleanup(res)

	indexes, drop := records.Poll(idx, p)
	recs := dataPoll(indexes, res)
	items, _ := tcode.Transcode(recs)

	_ = drop
	if params.Local {
		// setHeaders(c, m)
		items <- m
		if params.Format == "xml" && params.Fields != "" {
			fields := strings.Split(params.Fields, sepSign)
			streamSmplRecords(c, enc, items, fields)
		} else {
			streamAllRecords(c, enc, items)
		}
	} else {
		c.Stream(func(w io.Writer) bool {
			prms := &UrlParams{}
			prms.SetHost("52.3.59.171", "8765")
			prms.Path = "search"
			prms.Params = map[string]interface{}{
				"query":       params.Query,
				"files":       createFilesQuery(params.Files),
				"surrounding": params.Surrounding,
				"format":      params.Format,
				"fuzziness":   params.Fuzziness,
				"local":       true,
			}

			url := createClusterUrl(prms)
			response, err := http.Get(url)
			if err != nil {
				fmt.Printf("%s", err)
				c.JSON(500, err)
				return true
			}
			defer response.Body.Close()
			for k := range response.Header {
				c.Header(k, response.Header.Get(k))
				// fmt.Printf("HEADER %v : %v", k, v)
			}
			io.Copy(w, response.Body)
			return false
		})
	}
}

func logErrors(format string, errors chan error) {
	for err := range errors {
		if err != nil {
			log.Printf(format, err.Error())
		}
	}
}

func streamAllRecords(c *gin.Context, enc encoder.Encoder, recs chan interface{}) {
	first := true
	c.Stream(func(w io.Writer) bool {
		if first {
			enc.Begin(w)
			first = false
		}

		if record, ok := <-recs; ok {
			if err := enc.Write(w, record); err != nil {
				log.Panicln(err)
			} else {
				c.Writer.Flush()
			}
			return true
		}
		enc.End(w)
		return false

	})
}

func streamSmplRecords(c *gin.Context, enc encoder.Encoder, recs chan interface{}, sample []string) {
	first := true

	c.Stream(func(w io.Writer) bool {
		if first {
			enc.Begin(w)
			first = false
		}

		if record, ok := <-recs; ok {

			rec := map[string]interface{}{}

			for i := range sample {
				value, ok := record.(map[string]interface{})[sample[i]]
				if ok {
					rec[sample[i]] = value
				}
			}
			if err := enc.Write(w, rec); err != nil {
				log.Panicln(err)
			} else {
				c.Writer.Flush()
			}

			return true

		}
		enc.End(w)
		return false
	})
}

const (
	// PollingInterval is a const for polling time
	PollingInterval = time.Millisecond * 50
	// PollBufferCapacity is a max buffer size
	PollBufferCapacity = 64
)

func dataPoll(input chan records.IdxRecord, dataFile *os.File) chan records.IdxRecord {
	output := make(chan records.IdxRecord, PollBufferCapacity)
	go func() {
		for rec := range input {
			rec.Data = nextData(dataFile, rec.Length)
			output <- rec
		}
		close(output)
	}()
	return output
}

func nextData(res *os.File, length uint16) (result []byte) {
	var total uint16
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
