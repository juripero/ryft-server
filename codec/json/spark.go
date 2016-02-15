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

/* Spark JSON encoder uses an JSON array to store all records.
 Statistics and errors are ignored.

[ <records> ]
*/

// Spark JSON encoder.
type SparkEncoder struct {
	writer  io.Writer
	encoder *backend.Encoder

	records int // number of records written
}

// Create new Spark JSON encoder instance.
func NewSparkEncoder(w io.Writer) (*SparkEncoder, error) {
	enc := new(SparkEncoder)
	enc.encoder = backend.NewEncoder(w)
	enc.writer = w
	return enc, nil
}

// Write a RECORD
func (enc *SparkEncoder) EncodeRecord(rec interface{}) error {
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
		return err
	}

	enc.records += 1
	return nil // OK
}

// Write a STATISTICS
func (enc *SparkEncoder) EncodeStat(stat interface{}) error {
	return nil // OK, ignored
}

// Write an ERROR
func (enc *SparkEncoder) EncodeError(err error) error {
	return nil // OK, ignored
}

// End writing, close JSON object.
func (enc *SparkEncoder) Close() error {
	// write header for the first record
	if enc.records == 0 {
		err := enc.writeHeader()
		if err != nil {
			return err
		}
	}

	// end of records
	_, err := enc.writer.Write([]byte("]"))
	if err != nil {
		return err
	}

	return nil // OK
}

// Write JSON object header.
func (enc *SparkEncoder) writeHeader() error {
	_, err := enc.writer.Write([]byte{'['})
	return err
}
