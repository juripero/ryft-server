/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2018, Ryft Systems, Inc.
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
	backend "encoding/csv"
	"fmt"
	"io"
)

/* Stream CSV encoder uses tag prefixes for each item written

rec,<record>
err,<error message>
stat,<statistics>
end
*/
type StreamEncoder struct {
	encoder *backend.Writer
}

const (
	TAG_REC  = "rec"
	TAG_ERR  = "err"
	TAG_STAT = "stat"
	TAG_EOF  = "end"
)

// NewStreamEncoder creates new CSV stream encoder.
func NewStreamEncoder(w io.Writer) (*StreamEncoder, error) {
	enc := new(StreamEncoder)
	enc.encoder = backend.NewWriter(w)
	enc.encoder.Comma = ','
	return enc, nil
}

// write a CSV record
func (enc *StreamEncoder) encode(tag string, data interface{}) error {
	if data == nil {
		return nil
	}

	record := []string{tag}
	if i, ok := data.(Marshaler); ok {
		csv, err := i.MarshalCSV()
		if err != nil {
			return err
		}
		record = append(record, csv...)
	} else {
		// fallback: as a simple string
		record = append(record, fmt.Sprintf("%s", data))
	}

	return enc.encoder.Write(record)
}

// Write a RECORD
func (enc *StreamEncoder) EncodeRecord(data interface{}) error {
	return enc.encode(TAG_REC, data)
}

// Write a STATISTICS
func (enc *StreamEncoder) EncodeStat(data interface{}) error {
	return enc.encode(TAG_STAT, data)
}

// Write an ERROR
func (enc *StreamEncoder) EncodeError(err error) error {
	if err == nil {
		return nil
	}
	csv := []string{
		TAG_ERR,
		err.Error(),
	}
	return enc.encoder.Write(csv)
}

// End writing, close CSV object.
func (enc *StreamEncoder) Close() error {
	err := enc.encoder.Write([]string{TAG_EOF})
	if err != nil {
		return err
	}

	enc.encoder.Flush()
	return enc.encoder.Error()
}
