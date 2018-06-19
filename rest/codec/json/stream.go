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

package json

import (
	backend "encoding/json"
	"io"
)

/* Stream JSON encoder uses tag prefixes for each item written.

"rec" { <record> }
"rec" { <record> }
"err" "<error message>"
"rec" { <record> }
"stat" { <statistics> }
"end"
*/

// Stream JSON encoder.
type StreamEncoder struct {
	e *backend.Encoder // JSON encoder
}

const (
	TAG_REC  = "rec"
	TAG_ERR  = "err"
	TAG_STAT = "stat"
	TAG_EOF  = "end"
)

// Create new stream JSON encoder instance.
func NewStreamEncoder(w io.Writer) (*StreamEncoder, error) {
	enc := new(StreamEncoder)
	enc.e = backend.NewEncoder(w)
	return enc, nil
}

// Write a RECORD
func (enc *StreamEncoder) EncodeRecord(rec interface{}) error {
	if rec == nil {
		return nil // nothing to do
	}

	// write tag
	if err := enc.e.Encode(TAG_REC); err != nil {
		return err
	}

	// encode record
	if err := enc.e.Encode(rec); err != nil {
		return err
	}

	return nil // OK
}

// Write a STATISTICS
func (enc *StreamEncoder) EncodeStat(stat interface{}) error {
	if stat == nil {
		return nil // nothing to do
	}

	// write tag
	if err := enc.e.Encode(TAG_STAT); err != nil {
		return err
	}

	// encode statistics
	if err := enc.e.Encode(stat); err != nil {
		return err
	}

	return nil // OK
}

// Write an ERROR
func (enc *StreamEncoder) EncodeError(err_ error) error {
	if err_ == nil {
		return nil // nothing to do
	}

	// write tag
	if err := enc.e.Encode(TAG_ERR); err != nil {
		return err
	}

	// encode error as a string
	if err := enc.e.Encode(err_.Error()); err != nil {
		return err
	}

	return nil // OK
}

// End writing, close stream.
func (enc *StreamEncoder) Close() error {
	// write tag
	if err := enc.e.Encode(TAG_EOF); err != nil {
		return err
	}

	return nil // OK
}

// JSON stream decoder.
type StreamDecoder struct {
	d *backend.Decoder
}

// Create new stream JSON decoder instance.
func NewStreamDecoder(r io.Reader) (*StreamDecoder, error) {
	dec := new(StreamDecoder)
	dec.d = backend.NewDecoder(r)
	return dec, nil
}

// NextTag decodes next tag from the stream.
func (dec *StreamDecoder) NextTag() (string, error) {
	var tag string
	err := dec.d.Decode(&tag)
	if err != nil {
		return "", err
	}
	return tag, nil // OK
}

// Next decodes next item from the stream.
func (dec *StreamDecoder) Next(item interface{}) error {
	return dec.d.Decode(item)
}
