package csv

import (
	"testing"
	"bytes"
	"io"
	"github.com/stretchr/testify/assert"
	"errors"
	"fmt"
	"strconv"
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
	fmt.Println(bad)

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
