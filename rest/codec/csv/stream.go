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

package csv

import (
	"io"
	backend "encoding/csv"
	"strconv"

	"github.com/getryft/ryft-server/rest/format/utf8"
	"fmt"
	"log"
)

/* Stream CSV encoder uses tag prefixes for each item written

rec;<record>
err;<error message>
stat;<statistics>
end
*/
type StreamEncoder struct{
	encoder *backend.Writer
}

const (
	TAG_REC  = "rec"
	TAG_ERR  = "err"
	TAG_STAT = "stat"
	TAG_EOF  = "end"
)

func NewStreamEncoder(w io.Writer) (*StreamEncoder, error) {
	enc := new(StreamEncoder)
	enc.encoder = backend.NewWriter(w)
	enc.encoder.Comma = ','
	return enc, nil
}


func (enc *StreamEncoder) encode(tag string, data interface{}) error {
	if data == nil {
		return nil
	}
	record := []string{tag}
	// TODO: Replace it with the serializer. Looks weird.
	switch data := data.(type) {
	case *utf8.Record:
		//filename,offset,length,fuzziness,data
		tail := []string{
			data.Index.File,
			strconv.FormatUint(data.Index.Offset, 10),
			strconv.FormatUint(data.Index.Length,10),
			strconv.FormatInt(int64(data.Index.Fuzziness), 10),
			string(data.RawData),
		}
		record = append(record, tail...)
	case *utf8.Stat:
		// TODO: expand it
		// stat,total bytes,matches, etc...
		tail := []string{
			strconv.FormatUint(data.Matches, 10),
			strconv.FormatUint(data.TotalBytes, 10),
			strconv.FormatUint(data.Duration, 10),
			strconv.FormatFloat(data.DataRate, 'f', -1, 64),
			strconv.FormatUint(data.FabricDuration, 10),
			strconv.FormatFloat(data.FabricDataRate, 'f', -1, 64),
			data.Host,
		}
		record = append(record, tail...)
	case error:
		record = append(record, data.Error())
	case string: // only for tests.
		record = append(record, string(data))
	case int: // only for tests.
		record = append(record, strconv.Itoa(data))
	default:
		// for debug
		fmt.Printf("%#v", data)
	}
	if err := enc.encoder.Write(record); err != nil {
		return err
	}
	enc.encoder.Flush()
	if err := enc.encoder.Error(); err != nil {
		return err
	}
	return nil
}

// Write a RECORD
func (enc *StreamEncoder) EncodeRecord(rec interface{}) error {
	return enc.encode(TAG_REC, rec)
}

// Write a STATISTICS
func (enc *StreamEncoder) EncodeStat(stat interface{}) error {
	return enc.encode(TAG_STAT, stat)
}

// Write an ERROR
func (enc *StreamEncoder) EncodeError(err_ error) error {
	return enc.encode(TAG_ERR, err_)
}

// End writing, close CSV object.
func (enc *StreamEncoder) Close() error {
	if err := enc.encoder.Write([]string{TAG_EOF}); err != nil {
		return err
	}
	enc.encoder.Flush()
	if err := enc.encoder.Error(); err != nil {
		return err
	}
	return nil
}
