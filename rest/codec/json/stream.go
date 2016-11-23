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
	writer  io.Writer
	encoder *backend.Encoder
}

const (
	TAG_REC  = `"rec" `
	TAG_ERR  = `"err" `
	TAG_STAT = `"stat" `
	TAG_EOF  = `"end"`
)

// Create new stream JSON encoder instance.
func NewStreamEncoder(w io.Writer) (*StreamEncoder, error) {
	enc := new(StreamEncoder)
	enc.encoder = backend.NewEncoder(w)
	enc.writer = w
	return enc, nil
}

// Write a RECORD
func (enc *StreamEncoder) EncodeRecord(rec interface{}) error {
	// write tag
	_, err := enc.writer.Write([]byte(TAG_REC))
	if err != nil {
		return err
	}

	// encode record
	err = enc.encoder.Encode(rec)
	if err != nil {
		return err
	}

	return nil // OK
}

// Write a STATISTICS
func (enc *StreamEncoder) EncodeStat(stat interface{}) error {
	// write tag
	_, err := enc.writer.Write([]byte(TAG_STAT))
	if err != nil {
		return err
	}

	// encode statistics
	err = enc.encoder.Encode(stat)
	if err != nil {
		return err
	}

	return nil // OK
}

// Write an ERROR
func (enc *StreamEncoder) EncodeError(err_ error) error {
	// write tag
	_, err := enc.writer.Write([]byte(TAG_ERR))
	if err != nil {
		return err
	}

	// encode error as a string
	err = enc.encoder.Encode(err_.Error())
	if err != nil {
		return err
	}

	return nil // OK
}

// End writing, close stream.
func (enc *StreamEncoder) Close() error {
	// write tag
	_, err := enc.writer.Write([]byte(TAG_EOF))
	if err != nil {
		return err
	}

	return nil // OK
}

// JSON stream decoder.
type StreamDecoder struct {
	decoder *backend.Decoder
}

// Create new stream JSON decoder instance.
func NewStreamDecoder(r io.Reader) (*StreamDecoder, error) {
	dec := new(StreamDecoder)
	dec.decoder = backend.NewDecoder(r)
	return dec, nil
}

// NextTag decodes next tag from the stream.
func (dec *StreamDecoder) NextTag() (string, error) {
	var tag string
	err := dec.decoder.Decode(&tag)
	if err != nil {
		return "", err
	}
	return tag, nil // OK
}

// Next decodes next item from the stream.
func (dec *StreamDecoder) Next(item interface{}) error {
	return dec.decoder.Decode(item)
}
