package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test stream JSON encoder
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

			vals := []interface{}{}
			dec := json.NewDecoder(bytes.NewReader(buf.Bytes()))
			for {
				var val interface{}
				err := dec.Decode(&val)
				if err != nil {
					break
				}

				vals = append(vals, val)
			}

			sbuf, err := json.Marshal(vals)
			assert.NoError(t, err)
			assert.JSONEq(t, expected, string(sbuf))
		}
	}

	// test stream JSON encoder (bad case)
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
	}, `["end"]`)

	// one error
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(nil)) // ignored
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
	}, `["err", "err1", "end"]`)

	// a few errors
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err2")))
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err3")))
	}, `["err", "err1", "err", "err2", "err", "err3", "end"]`)

	// one record
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeRecord(nil)) // ignored
		assert.NoError(t, enc.EncodeRecord("rec1"))
	}, `["rec", "rec1", "end"]`)

	// a few records
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeRecord("rec1"))
		assert.NoError(t, enc.EncodeRecord("rec2"))
		assert.NoError(t, enc.EncodeRecord("rec3"))
	}, `["rec", "rec1", "rec", "rec2", "rec", "rec3", "end"]`)

	// a few records and error
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
		assert.NoError(t, enc.EncodeRecord("rec1"))
		assert.NoError(t, enc.EncodeRecord("rec2"))
		assert.NoError(t, enc.EncodeRecord("rec3"))
	}, `["err", "err1", "rec", "rec1", "rec", "rec2", "rec", "rec3", "end"]`)

	// stat
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeStat(nil)) // ignored
		assert.NoError(t, enc.EncodeStat(555))
	}, `["stat", 555, "end"]`)

	// bad cases
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeRecord("rec1"); err != nil {
			return err
		}
		return nil
	}, 4, "EOF") // EncodeRecord (tag)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeRecord("rec1"); err != nil {
			return err
		}
		return nil
	}, 8, "EOF") // EncodeRecord (record itself)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeStat("stat1"); err != nil {
			return err
		}
		return nil
	}, 4, "EOF") // EncodeStat (tag)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeStat("stat1"); err != nil {
			return err
		}
		return nil
	}, 8, "EOF") // EncodeStat (record itself)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeError(fmt.Errorf("err1")); err != nil {
			return err
		}
		return nil
	}, 4, "EOF") // EncodeStat (tag)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeError(fmt.Errorf("err1")); err != nil {
			return err
		}
		return nil
	}, 8, "EOF") // EncodeStat (record itself)
	bad(func(enc *StreamEncoder) error {
		return nil
	}, 2, "EOF") // Close
}

// Test stream JSON decoder
func TestStreamDecoder(t *testing.T) {
	// test stream JSON decoder
	check := func(data string, expected string) {
		dec, err := NewStreamDecoder(bytes.NewBufferString(data))
		if assert.NoError(t, err) {
			assert.NotNil(t, dec)

			vals := []interface{}{}
			for {
				tag, err := dec.NextTag()
				assert.NoError(t, err)
				vals = append(vals, tag)

				if tag == TAG_EOF {
					break
				}

				var val interface{}
				err = dec.Next(&val)
				assert.NoError(t, err)
				vals = append(vals, val)
			}

			sbuf, err := json.Marshal(vals)
			assert.NoError(t, err)
			assert.JSONEq(t, expected, string(sbuf))
		}
	}

	// test stream JSON decoder (bad cases)
	bad := func(data string, expectedError string) {
		dec, err := NewStreamDecoder(bytes.NewBufferString(data))
		if assert.NoError(t, err) {
			assert.NotNil(t, dec)

			var tag string
			for {
				tag, err = dec.NextTag()
				if err == nil {
					if tag == TAG_EOF {
						break
					}

					var val interface{}
					err = dec.Next(&val)
				}
				if err != nil {
					break
				}
			}

			if assert.Error(t, err) {
				assert.Contains(t, err.Error(), expectedError)
			}
		}
	}

	// empty
	check(
		`"end"`,
		`["end"]`)

	// one error
	check(
		`"err" "err1"
		 "end"`,
		`["err", "err1", "end"]`)

	// a few errors
	check(
		`"err" "err1"
		"err" "err2"
		"err" "err3"
		"end"`,
		`["err", "err1", "err", "err2", "err", "err3", "end"]`)

	// one record
	check(
		`"rec" "rec1"
		"end"`,
		`["rec", "rec1", "end"]`)

	// a few records
	check(
		`"rec" "rec1"
		"rec" "rec2"
		"rec" "rec3"
		"end"`,
		`["rec", "rec1", "rec", "rec2", "rec", "rec3", "end"]`)

	// a few records and error
	check(
		`"err" "err1"
		"rec" "rec1"
		"rec" "rec2"
		"rec" "rec3"
		"end"`,
		`["err", "err1", "rec", "rec1", "rec", "rec2", "rec", "rec3", "end"]`)

	// stat
	check(
		`"stat" 555
		"end"`,
		`["stat", 555, "end"]`)

	// bad cases
	bad(`"en`, "EOF")
}
