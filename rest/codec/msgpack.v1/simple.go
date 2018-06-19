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

package msgpack

import (
	"io"

	backend "github.com/ugorji/go/codec"
)

/* Simple MSGPACK encoder uses stream of RECORDS.
Errors and statistics are written at the end.
*/

// MSGPACK simple encoder.
type SimpleEncoder struct {
	RecordsOnly bool // do not write errors and statistics

	encoder *backend.Encoder
	errors  []string
	stat    interface{}
}

// Create new simple MSGPACK encoder instance.
func NewSimpleEncoder(w io.Writer) (*SimpleEncoder, error) {
	enc := new(SimpleEncoder)
	h := new(backend.MsgpackHandle)
	enc.encoder = backend.NewEncoder(w, h)
	return enc, nil
}

// Write a RECORD
func (enc *SimpleEncoder) EncodeRecord(rec interface{}) error {
	if rec == nil {
		return nil // nothing to do
	}

	return enc.encoder.Encode(rec)
}

// Write a STATISTICS
func (enc *SimpleEncoder) EncodeStat(stat interface{}) error {
	enc.stat = stat // will be written later
	return nil      // OK
}

// Write an ERROR
func (enc *SimpleEncoder) EncodeError(err_ error) error {
	if err_ != nil {
		// just save, will be written later
		enc.errors = append(enc.errors, err_.Error())
	}

	return nil // OK
}

// End writing, close JSON object.
func (enc *SimpleEncoder) Close() error {
	// write errors
	for _, emsg := range enc.errors {
		err := enc.encoder.Encode(emsg)
		if err != nil {
			return err
		}
	}

	// write statistics
	if enc.stat != nil {
		err := enc.encoder.Encode(enc.stat)
		if err != nil {
			return err
		}
	}

	return nil // OK
}
