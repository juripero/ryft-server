package json

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test simple JSON encoder
func TestSimpleEncoder(t *testing.T) {
	// test simple JSON encoder
	check := func(populate func(enc *SimpleEncoder), expected string) {
		buf := &bytes.Buffer{}
		enc, err := NewSimpleEncoder(buf)
		if assert.NoError(t, err) {
			assert.NotNil(t, enc)
			populate(enc)
			err = enc.Close()
			assert.NoError(t, err)

			assert.JSONEq(t, expected, buf.String())
		}
	}

	// test simple JSON encoder (bad case)
	bad := func(populate func(enc *SimpleEncoder) error, limit int, expectedError string) {
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

	// empty
	check(func(enc *SimpleEncoder) {
		// do nothing
	}, `{"results":[]}`)

	// one error
	check(func(enc *SimpleEncoder) {
		assert.NoError(t, enc.EncodeError(nil))
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
	}, `{"results":[], "errors":["err1"]}`)

	// a few errors
	check(func(enc *SimpleEncoder) {
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err2")))
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err3")))
	}, `{"results":[], "errors":["err1","err2","err3"]}`)

	// one record
	check(func(enc *SimpleEncoder) {
		assert.NoError(t, enc.EncodeRecord(nil)) // ignored
		assert.NoError(t, enc.EncodeRecord("rec1"))
	}, `{"results":["rec1"]}`)

	// a few records
	check(func(enc *SimpleEncoder) {
		assert.NoError(t, enc.EncodeRecord("rec1"))
		assert.NoError(t, enc.EncodeRecord("rec2"))
		assert.NoError(t, enc.EncodeRecord("rec3"))
	}, `{"results":["rec1", "rec2", "rec3"]}`)

	// a few records and error
	check(func(enc *SimpleEncoder) {
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
		assert.NoError(t, enc.EncodeRecord("rec1"))
		assert.NoError(t, enc.EncodeRecord("rec2"))
		assert.NoError(t, enc.EncodeRecord("rec3"))
	}, `{"results":["rec1", "rec2", "rec3"], "errors":["err1"]}`)

	// stat
	check(func(enc *SimpleEncoder) {
		assert.NoError(t, enc.EncodeStat(nil)) // ignored
		assert.NoError(t, enc.EncodeStat(555))
	}, `{"results":[], "stats":555}`)

	// bad cases
	bad(func(enc *SimpleEncoder) error {
		if err := enc.EncodeRecord("rec1"); err != nil {
			return err
		}
		return nil
	}, 10, "EOF") // EncodeRecord (header)
	bad(func(enc *SimpleEncoder) error {
		if err := enc.EncodeRecord("rec1"); err != nil {
			return err
		}
		return nil
	}, 14, "EOF") // EncodeRecord (record itself)
	bad(func(enc *SimpleEncoder) error {
		if err := enc.EncodeRecord("rec1"); err != nil {
			return err
		}
		if err := enc.EncodeRecord("rec2"); err != nil {
			return err
		}
		return nil
	}, 19, "EOF") // EncodeRecord (coma between)
	bad(func(enc *SimpleEncoder) error {
		// do nothing
		return nil
	}, 10, "EOF")
	bad(func(enc *SimpleEncoder) error {
		// do nothing
		return nil
	}, 12, "EOF")
	bad(func(enc *SimpleEncoder) error {
		if err := enc.EncodeError(fmt.Errorf("err1")); err != nil {
			return err
		}
		return nil
	}, 18, "EOF") // Close (error)
	bad(func(enc *SimpleEncoder) error {
		if err := enc.EncodeError(fmt.Errorf("err1")); err != nil {
			return err
		}
		return nil
	}, 27, "EOF") // Close (error)
	bad(func(enc *SimpleEncoder) error {
		if err := enc.EncodeStat("stat1"); err != nil {
			return err
		}
		return nil
	}, 18, "EOF") // Close (error)
	bad(func(enc *SimpleEncoder) error {
		if err := enc.EncodeStat("stat1"); err != nil {
			return err
		}
		return nil
	}, 24, "EOF") // Close (error)
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
