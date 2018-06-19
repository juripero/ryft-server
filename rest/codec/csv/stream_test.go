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
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
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
	"bytes"
	"errors"
	"io"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Int int

func (i Int) MarshalCSV() ([]string, error) {
	return []string{
		strconv.Itoa(int(i)),
	}, nil
}

type Record string

func (rec Record) MarshalCSV() ([]string, error) {
	return []string{
		string(rec),
	}, nil
}

// Test stream CSV encoder
func TestStreamEncoder(t *testing.T) {
	// test stream JSON encoder
	check := func(populate func(enc *StreamEncoder), expected string) {
		buf := &bytes.Buffer{}
		enc, err := NewStreamEncoder(buf)
		if assert.NoError(t, err) {
			assert.NotNil(t, enc)
			populate(enc)
			err = enc.Close()
			assert.NoError(t, err)
			assert.EqualValues(t, expected, buf.String())
		}
	}
	// test stream encoder (bad case)
	bad := func(populate func(enc *StreamEncoder) error, limit int, expectedError string) {
		buf := &bytes.Buffer{}
		lim := newLimitedWriter(buf, limit)
		enc, err := NewStreamEncoder(lim)
		if assert.NoError(t, err) {
			assert.NotNil(t, enc)
			err := populate(enc)
			if err == nil {
				err = enc.Close()
			}

			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), expectedError)
			}
		}
	}

	// empty
	check(func(enc *StreamEncoder) {
		// do nothing
	}, "end\n")
	// one error
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(nil)) // ignored
		assert.NoError(t, enc.EncodeError(errors.New("err1")))
	}, "err,err1\nend\n")
	// a few errors
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(errors.New("err1")))
		assert.NoError(t, enc.EncodeError(errors.New("err2")))
		assert.NoError(t, enc.EncodeError(errors.New("err3")))
	}, "err,err1\nerr,err2\nerr,err3\nend\n")
	// one record
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeRecord(nil)) // ignored
		assert.NoError(t, enc.EncodeRecord(Record(Record("rec1"))))
	}, "rec,rec1\nend\n")

	// a few records
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeRecord(Record("rec1")))
		assert.NoError(t, enc.EncodeRecord(Record("rec2")))
		assert.NoError(t, enc.EncodeRecord(Record("rec3")))
	}, "rec,rec1\nrec,rec2\nrec,rec3\nend\n")

	// a few records and error
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(errors.New("err1")))
		assert.NoError(t, enc.EncodeRecord(Record("rec1")))
		assert.NoError(t, enc.EncodeRecord(Record("rec2")))
		assert.NoError(t, enc.EncodeRecord(Record("rec3")))
	}, "err,err1\nrec,rec1\nrec,rec2\nrec,rec3\nend\n")

	// stat
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeStat(nil)) // ignored
		assert.NoError(t, enc.EncodeStat(Int(555)))
	}, "stat,555\nend\n")
	// bad cases
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeRecord(Record("rec1")); err != nil {
			return err
		}
		return nil
	}, 4, "") // EncodeRecord (tag)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeRecord(Record("rec1")); err != nil {
			return err
		}
		return nil
	}, 8, "") // EncodeRecord (record itself)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeStat("stat1"); err != nil {
			return err
		}
		return nil
	}, 4, "") // EncodeStat (tag)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeStat("stat1"); err != nil {
			return err
		}
		return nil
	}, 8, "") // EncodeStat (record itself)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeError(errors.New("err1")); err != nil {
			return err
		}
		return nil
	}, 4, "") // EncodeStat (tag)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeError(errors.New("err1")); err != nil {
			return err
		}
		return nil
	}, 8, "") // EncodeStat (record itself)
	bad(func(enc *StreamEncoder) error {
		return nil
	}, 2, "EOF") // Close

}

// limited writer
type limitedWriter struct {
	w io.Writer
	n int
}

// create new limited writer
func newLimitedWriter(w io.Writer, n int) *limitedWriter {
	return &limitedWriter{
		w: w,
		n: n,
	}
}

// Write
func (w *limitedWriter) Write(p []byte) (n int, err error) {
	if w.n <= 0 {
		return 0, io.EOF
	}

	if w.n < len(p) {
		n, err = w.w.Write(p[0:w.n])
		if err == nil {
			err = io.EOF
		}
	} else {
		n, err = w.w.Write(p)
	}

	w.n -= n
	return
}
