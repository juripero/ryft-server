package csv


import (
	"testing"
	"bytes"
	"io"

	"github.com/stretchr/testify/assert"
	"fmt"
	"errors"
)

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
	fmt.Print(bad)
	// empty
	check(func(enc *StreamEncoder) {
		// do nothing
	}, "")
	// one error
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(nil)) // ignored
		assert.NoError(t, enc.EncodeError(errors.New("err1")))
	}, "err,err1\r\n")
	// a few errors
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(errors.New("err1")))
		assert.NoError(t, enc.EncodeError(errors.New("err2")))
		assert.NoError(t, enc.EncodeError(errors.New("err3")))
	}, "err,err1\r\nerr,err2\r\nerr,err3\r\n")
	// one record
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeRecord(nil)) // ignored
		assert.NoError(t, enc.EncodeRecord("rec1"))
	}, "rec,rec1\r\n")

	// a few records
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeRecord("rec1"))
		assert.NoError(t, enc.EncodeRecord("rec2"))
		assert.NoError(t, enc.EncodeRecord("rec3"))
	}, "rec,rec1\r\nrec,rec2\r\nrec,rec3\r\n")

	// a few records and error
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(errors.New("err1")))
		assert.NoError(t, enc.EncodeRecord("rec1"))
		assert.NoError(t, enc.EncodeRecord("rec2"))
		assert.NoError(t, enc.EncodeRecord("rec3"))
	}, "err,err1\r\nrec,rec1\r\nrec,rec2\r\nrec,rec3\r\n")

	// stat
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeStat(nil)) // ignored
		assert.NoError(t, enc.EncodeStat(555))
	}, "stat,555\r\n")
	// bad cases
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeRecord("rec1"); err != nil {
			return err
		}
		return nil
	}, 4, "") // EncodeRecord (tag)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeRecord("rec1"); err != nil {
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
