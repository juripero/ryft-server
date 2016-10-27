package msgpack

// TODO: synchronize these tests with v1

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test simple encoder
func testSimpleEncoder(t *testing.T, populate func(enc *SimpleEncoder), expected string) {
	buf := &bytes.Buffer{}
	enc, err := NewSimpleEncoder(buf)
	if assert.NoError(t, err) {
		assert.NotNil(t, enc)
		populate(enc)
		err = enc.Close()
		assert.NoError(t, err)

		assert.EqualValues(t, expected, buf.String())
	}
}

// test simple encoder (bad case)
func testSimpleEncoderBad(t *testing.T, populate func(enc *SimpleEncoder) error, limit int, expectedError string) {
	buf := &bytes.Buffer{}
	lim := newLimitedWriter(buf, limit)
	enc, err := NewSimpleEncoder(lim)
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

// Test simple encoder, TODO: review these tests
func TestSimpleEncoder(t *testing.T) {
	_ = fmt.Print
	// empty
	testSimpleEncoder(t, func(enc *SimpleEncoder) {
		// do nothing
	}, "")

	// one error
	testSimpleEncoder(t, func(enc *SimpleEncoder) {
		assert.NoError(t, enc.EncodeError(nil))
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
	}, "\xa4err1")

	// a few errors
	testSimpleEncoder(t, func(enc *SimpleEncoder) {
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err2")))
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err3")))
	}, "\xa4err1\xa4err2\xa4err3")

	// one record
	testSimpleEncoder(t, func(enc *SimpleEncoder) {
		assert.NoError(t, enc.EncodeRecord(nil)) // ignored
		assert.NoError(t, enc.EncodeRecord("rec1"))
	}, "\xa4rec1")

	// a few records
	testSimpleEncoder(t, func(enc *SimpleEncoder) {
		assert.NoError(t, enc.EncodeRecord("rec1"))
		assert.NoError(t, enc.EncodeRecord("rec2"))
		assert.NoError(t, enc.EncodeRecord("rec3"))
	}, "\xa4rec1\xa4rec2\xa4rec3")

	// a few records and error
	testSimpleEncoder(t, func(enc *SimpleEncoder) {
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
		assert.NoError(t, enc.EncodeRecord("rec1"))
		assert.NoError(t, enc.EncodeRecord("rec2"))
		assert.NoError(t, enc.EncodeRecord("rec3"))
	}, "\xa4rec1\xa4rec2\xa4rec3\xa4err1")

	// stat
	testSimpleEncoder(t, func(enc *SimpleEncoder) {
		assert.NoError(t, enc.EncodeStat(nil)) // ignored
		assert.NoError(t, enc.EncodeStat(555))
	}, "\xcd\x02+")

	// bad cases
	testSimpleEncoderBad(t, func(enc *SimpleEncoder) error {
		if err := enc.EncodeRecord("rec1"); err != nil {
			return err
		}
		return nil
	}, 4, "EOF")
	testSimpleEncoderBad(t, func(enc *SimpleEncoder) error {
		if err := enc.EncodeError(fmt.Errorf("err1")); err != nil {
			return err
		}
		return nil
	}, 4, "EOF") // Close (error)
	testSimpleEncoderBad(t, func(enc *SimpleEncoder) error {
		if err := enc.EncodeStat("stat1"); err != nil {
			return err
		}
		return nil
	}, 4, "EOF") // Close (error)
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
