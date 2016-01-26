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
	"github.com/getryft/ryft-server/records"
	"github.com/getryft/ryft-server/search"
)

type RawTranscoder struct {
	Transcoder
}

type RawData struct {
	Index Index       `json:"_index"`
	Data  interface{} `json:"data"`
}

func (transcoder *RawTranscoder) Transcode(recs chan records.IdxRecord) (chan interface{}, chan error) {
	output := make(chan interface{}, TranscodeBufferCapacity)
	errors := make(chan error)

	go func() {
		defer close(output)
		defer close(errors)
		for rec := range recs {
			output <- RawData{
				Index{rec.File, rec.Offset, rec.Length, rec.Fuzziness},
				rec.Data,
			}
		}
	}()

	return output, errors
}

func (transcoder *RawTranscoder) Transcode1(rec *search.Record) (interface{}, error) {
	return RawData{Index: NewIndex(rec.Index), Data: rec.Data}, nil
}

func DecodeRawItem(item *RawData) (*search.Record, error) {
	return &search.Record{
		Index: search.Index{
			File:      item.Index.File,
			Offset:    item.Index.Offset,
			Length:    uint64(item.Index.Length),
			Fuzziness: item.Index.Fuzziness,
		},
		Data: item.Data,
	}, nil
}

func DecodeRawStat(stat *Statistics) (search.Statistics, error) {
	return search.Statistics{
		Matches:    stat.Matches,
		TotalBytes: stat.TotalBytes,
		Duration:   stat.Duration,
	}, nil
}

func (transcoder *RawTranscoder) TranscodeStat(stat search.Statistics) (interface{}, error) {
	return NewStat(stat), nil
}
