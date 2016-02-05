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

/* Simple JSON encoder uses an JSON object to store all data.
 Statistics and errors are cached by encoder till the end.

{ results: [ <records> ]
, errors: [ <errors ]
, stats: <statistics>
}
*/

// Simple JSON encoder.
type SimpleEncoder struct {
	writer  io.Writer
	encoder *backend.Encoder

	records int         // number of records written
	stat    interface{} // cached statistics
	errors  []string    // cached error messages
}

// Create new simple JSON encoder instance.
func NewSimpleEncoder(w io.Writer) (*SimpleEncoder, error) {
	enc := new(SimpleEncoder)
	enc.encoder = backend.NewEncoder(w)
	enc.writer = w
	return enc, nil
}

// Write a RECORD
func (enc *SimpleEncoder) EncodeRecord(rec interface{}) error {
	// write header for the first record
	if enc.records == 0 {
		err := enc.writeHeader()
		if err != nil {
			return err
		}
	}

	// write coma separator for all
	// records except the first one
	if enc.records > 0 {
		_, err := enc.writer.Write([]byte{','})
		if err != nil {
			return err
		}
	}

	// encode record
	err := enc.encoder.Encode(rec)
	if err != nil {
		return nil
	}

	enc.records += 1
	return nil // OK
}

// Write a STATISTICS
func (enc *SimpleEncoder) EncodeStat(stat interface{}) error {
	enc.stat = stat // will be written later
	return nil      // OK
}

// Write an ERROR
func (enc *SimpleEncoder) EncodeError(err error) error {
	if err != nil {
		// just save, will be written later
		enc.errors = append(enc.errors, err.Error())
	}

	return nil // OK
}

// End writing, close JSON object.
func (enc *SimpleEncoder) Close() error {
	// write header for the first record
	if enc.records == 0 {
		err := enc.writeHeader()
		if err != nil {
			return err
		}
	}

	// array of errors
	if len(enc.errors) > 0 {
		// write "errors" header
		_, err := enc.writer.Write([]byte(`,"errors":`))
		if err != nil {
			return err
		}

		// write errors
		err = enc.encoder.Encode(enc.errors)
		if err != nil {
			return err
		}
	}

	// statistics
	if enc.stat != nil {
		// write "stats" header
		_, err := enc.writer.Write([]byte(`,"stats":`))
		if err != nil {
			return err
		}

		// write statistics
		err = enc.encoder.Encode(enc.stat)
		if err != nil {
			return err
		}
	}

	// write footer
	_, err := enc.writer.Write([]byte("}"))
	return err
}

// Write JSON object header.
func (enc *SimpleEncoder) writeHeader() error {
	_, err := enc.writer.Write([]byte(`{"results":[`))
	return err
}
