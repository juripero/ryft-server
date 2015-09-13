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

	"github.com/getryft/ryft-rest-api/rol"
	"github.com/getryft/ryft-rest-api/ryft-server/encoder"
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


func search(c *gin.Context) {
	defer srverr.DeferRecover(c)

	var err error

	// parse request parameters
	var params SearchParams
	if err = c.Bind(&params); err != nil {
		panic(srverr.New(http.StatusBadRequest, err.Error()))
	}

	// get a new unique search index
	n := names.New()
	log.Printf("SEARCH(%d): %s", n.Index, c.Request.URL.String())

//	log.Printf("** start binding")
//	var s *binding.Search
//	var err error
//	if s, err = binding.NewSearch(c); err != nil {
//		panic(srverr.New(http.StatusBadRequest, err.Error()))
//	}
//	c.Header("Content-Type", gin.MIMEPlain)
//	if s.IsOutJson() {
//		c.Header("Content-Type", gin.MIMEPlain)
//	} else if s.IsOutMsgpk() {
//		c.Header("Content-Type", "application/x-msgpack")
//	} else {
//		panic(srverr.New(http.StatusBadRequest, "Supported formats (Content-Type): application/json, application/x-msgpack"))
//	}

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

	recs, drop := records.Poll(idx, p)

	_ = drop

	accept := c.NegotiateFormat(encoder.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = encoder.MIMEJSON
	}
	log.Printf("ACCEPT(%d): %s", n.Index, accept)

	var enc encoder.Encoder
	if enc, err = encoder.GetByMimeType(accept); err != nil {
		panic(srverr.New(http.StatusBadRequest, err.Error()))
	}

	c.Header("Content-Type", accept)

	first := true

	c.Stream(func(w io.Writer) bool {

		if first {
			enc.Begin(w)
			first = false
		}

		if record, ok := <-recs; ok {
			log.Printf("RECORD: %v", record)
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

//		switch s.State {
//			case binding.StateBegin:
//				log.Println("StateBegin")
//				s.State = binding.StateBody
//			case binding.StateBody:
//				log.Println("StateBody")
//				s.State = binding.StateEnd
//			case binding.StateEnd:
//				log.Println("StateEnd")
//		}
//		err = outstream.Write(s, recs, res, w, drop)
//		if err != nil {
//			idx.Close()
//			idx = nil
//			if !KeepResults {
//				os.Remove(names.ResultsDirPath(n.IdxFile))
//			}
//			res.Close()
//			res = nil
//			if !KeepResults {
//				os.Remove(names.ResultsDirPath(n.ResultFile))
//			}
//		}
//		return false
	})
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
		resultsDs := func() *rol.RolDS {
			return ds.SearchFuzzyHamming(names.PathInRyftoneForResultDir(n.ResultFile), s.Query, s.Surrounding, s.Fuzziness, "", &idxFile, s.CaseSensitive)
		}()
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
