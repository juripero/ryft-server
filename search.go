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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/getryft/ryft-server/encoder"
	"github.com/getryft/ryft-server/names"
	"github.com/getryft/ryft-server/ryftprim"
	"github.com/getryft/ryft-server/srverr"
	"github.com/getryft/ryft-server/transcoder"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
)

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
	Nodes         uint8    `form:"nodes" json:"nodes"`
	Local         bool     `form:"local" json:"local"`
	Stats         bool     `form:"stats" json:"stats"`
}

func search(c *gin.Context) {

	defer srverr.Recover(c)

	var err error

	// parse request parameters
	params := SearchParams{}
	params.Format = transcoder.RAWTRANSCODER
	if err = c.Bind(&params); err != nil {
		panic(srverr.New(http.StatusBadRequest, err.Error()))
	}

	// setting up transcoder to convert raw data
	var tcode transcoder.Transcoder
	if tcode, err = transcoder.GetByFormat(params.Format); err != nil {
		panic(srverr.New(http.StatusBadRequest, err.Error()))
	}

	enc := encoder.FromContext(c)

	// get a new unique search index
	n := names.New()
	log.Printf("SEARCH(%d): %s", n.Index, c.Request.URL.String())

	query, aErr := url.QueryUnescape(params.Query)

	if aErr != nil {
		panic(srverr.New(http.StatusBadRequest, aErr.Error()))
	}

	results := ryftprim.Search(&ryftprim.Params{
		Query:         query,
		Files:         params.Files,
		Surrounding:   params.Surrounding,
		Fuzziness:     params.Fuzziness,
		Format:        params.Format,
		CaseSensitive: params.CaseSensitive,
		Nodes:         params.Nodes,
		IndexFile:     n.FullIndexPath(),
		ResultsFile:   n.FullResultsPath(),
		KeepFiles:     *KeepResults,
	})

	// TODO: for cloud code get other ryftprim.Result objects and merge together
	// [[[ ]]]]

	items, _ := tcode.Transcode(results.Results)

	//	if params.Local {
	if !params.Stats {
		results.Stats = nil
	}
	// if params.Format == "xml" && params.Fields != "" {
	// fields := strings.Split(params.Fields, sepSign)
	// streamSmplRecords(c, enc, results, fields)
	// } else {
	streamAllRecords(c, enc, items, results.Stats)
	// }
	//	} else {
	//		cnslSrvc, err := GetConsulInfo()
	//		if err != nil {
	//			panic(srverr.New(http.StatusInternalServerError, err.Error()))
	//		}
	//		ch := merge(items)
	//		// time.Sleep(10 * time.Second)
	//		for _, srv := range cnslSrvc {
	//			recsChan, _ := searchInNode(params, srv)
	//			ch = merge(ch, recsChan)
	//		}
	//		streamAllRecords(c, enc, ch, nil)
	//	}
}

func searchInNode(params SearchParams, service *api.CatalogService) (recs chan interface{}, chanErr chan error) {
	recs = make(chan interface{})
	chanErr = make(chan error)

	go func() {
		if compareIP(service.Address) && service.ServicePort == (*listenAddress).Port {
			close(recs)
			close(chanErr)
			return
		}

		prms := &UrlParams{}
		prms.SetHost(service.ServiceAddress, fmt.Sprint(service.ServicePort))
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
			close(recs)
			chanErr <- err
			return
		}

		defer response.Body.Close()
		dec := json.NewDecoder(response.Body)
		var v map[string][]map[string]interface{}
		dec.Decode(&v)

		if i, ok := v["results"]; ok {
			for _, v := range i {
				if index, ok := v["_index"]; ok {
					index.(map[string]interface{})["address"] = prms.host
				}
			}
		}

		recs <- v

		// m := map[string]string{
		// "address": prms.host,
		// }
		// recs <- m
		// recs <- response.Body
		defer close(chanErr)
		defer close(recs)
	}()
	return
}

func logErrors(format string, errors chan error) {
	for err := range errors {
		if err != nil {
			log.Printf(format, err.Error())
		}
	}
}

func streamAllRecords(c *gin.Context, enc encoder.Encoder, results chan interface{}, stats chan ryftprim.Statistics) {

	first := true
	c.Stream(func(w io.Writer) bool {
		if first {
			enc.Begin(w)
			first = false
		}

		if record, ok := <-results; ok {
			if err := enc.Write(w, record); err != nil {
				log.Panicln(err)
			} else {
				c.Writer.Flush()
			}
			return true
		}
		if stats != nil {
			s := <-stats
			enc.EndWithStats(w, s.AsMap())
		} else {
			enc.End(w)
		}
		return false

	})
}

// func streamSmplRecords(c *gin.Context, enc encoder.Encoder, result *ryftprim.Result, sample []string) {
// first := true
//
// c.Stream(func(w io.Writer) bool {
// 	if first {
// 		enc.Begin(w)
// 		first = false
// 	}
//
// 	if record, ok := <-result.Results; ok {
//
// 		rec := map[string]interface{}{}
//
// 		for i := range sample {
// 			value, ok := record.(map[string]interface{})[sample[i]]
// 			if ok {
// 				rec[sample[i]] = value
// 			}
// 		}
// 		if err := enc.Write(w, rec); err != nil {
// 			log.Panicln(err)
// 		} else {
// 			c.Writer.Flush()
// 		}
//
// 		return true
//
// 	}
//
// 	if result.Stats != nil {
// 		stats := <-result.Stats
// 		enc.EndWithStats(w, stats.AsMap())
// 	} else {
// 		enc.End(w)
// 	}
// 	return false
// })
// }
