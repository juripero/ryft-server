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

package transcoder

import (
	"errors"
	"fmt"

	"github.com/clbanning/mxj"
	"github.com/getryft/ryft-server/search"
)

type XmlTranscoder struct {
	Transcoder
}

func (transcoder *XmlTranscoder) Transcode1(rec *search.Record, fields []string) (res interface{}, err error) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in parsing ", r)
			//          debug.PrintStack()
			//          log.Printf("PASRING XML: %s", rec.Data)
			err = errors.New(fmt.Sprintf("PASRING XML: %s", rec))
			return
		}
	}()

	obj, err := mxj.NewMapXml(rec.Data.([]byte))
	if err != nil {
		return nil, err
	}
	for k := range obj {
		item, ok := obj[k]
		if ok {
			switch i := item.(type) {
			case map[string]interface{}:
				// if fields is not empty - do filtering
				if len(fields) != 0 {
					res = make(map[string]interface{})
					for _, k := range fields {
						if r, ok := i[k]; ok {
							res.(map[string]interface{})[k] = r
						}
					}
				} else {
					res = i
				}
				res.(map[string]interface{})["_index"] = NewIndex(rec.Index)
				break
			default:
				return nil, fmt.Errorf("incorrect input data type")
			}
		}
		break
	}
	return
}

func (transcoder *XmlTranscoder) TranscodeStat(stat *search.Statistics) (interface{}, error) {
	// TODO: replace with XML?
	return NewStat(stat), nil
}
