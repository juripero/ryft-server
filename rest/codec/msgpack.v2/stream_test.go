package msgpack

// TODO: synchronize these tests with v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test stream encoder
func TestStreamEncoder(t *testing.T) {
	// test stream encoder
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
	}, "\x00")

	// one error
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(nil)) // ignored
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
	}, "\x02\xa4err1\x00")

	// a few errors
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err2")))
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err3")))
	}, "\x02\xa4err1\x02\xa4err2\x02\xa4err3\x00")

	// one record
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeRecord(nil)) // ignored
		assert.NoError(t, enc.EncodeRecord("rec1"))
	}, "\x01\xa4rec1\x00")

	// a few records
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeRecord("rec1"))
		assert.NoError(t, enc.EncodeRecord("rec2"))
		assert.NoError(t, enc.EncodeRecord("rec3"))
	}, "\x01\xa4rec1\x01\xa4rec2\x01\xa4rec3\x00")

	// a few records and error
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
		assert.NoError(t, enc.EncodeRecord("rec1"))
		assert.NoError(t, enc.EncodeRecord("rec2"))
		assert.NoError(t, enc.EncodeRecord("rec3"))
	}, "\x02\xa4err1\x01\xa4rec1\x01\xa4rec2\x01\xa4rec3\x00")

	// stat
	check(func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeStat(nil)) // ignored
		assert.NoError(t, enc.EncodeStat(555))
	}, "\x03\xcd\x02+\x00")

	// bad cases
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeRecord("rec1"); err != nil {
			return err
		}
		return nil
	}, 0, "EOF") // EncodeRecord (tag)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeRecord("rec1"); err != nil {
			return err
		}
		return nil
	}, 4, "EOF") // EncodeRecord (record itself)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeStat("stat1"); err != nil {
			return err
		}
		return nil
	}, 0, "EOF") // EncodeStat (tag)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeStat("stat1"); err != nil {
			return err
		}
		return nil
	}, 4, "EOF") // EncodeStat (record itself)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeError(fmt.Errorf("err1")); err != nil {
			return err
		}
		return nil
	}, 0, "EOF") // EncodeStat (tag)
	bad(func(enc *StreamEncoder) error {
		if err := enc.EncodeError(fmt.Errorf("err1")); err != nil {
			return err
		}
		return nil
	}, 4, "EOF") // EncodeStat (record itself)
	bad(func(enc *StreamEncoder) error {
		return nil
	}, 0, "EOF") // Close
}

// Test stream decoder
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
				switch v := val.(type) {
				case []byte:
					vals = append(vals, string(v))
				default:
					vals = append(vals, val)

				}
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

			var tag Tag
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
		"\x00",
		`[0]`)

	// one error
	check(
		"\x02\xa4err1\x00",
		`[2, "err1", 0]`)

	// a few errors
	check(
		"\x02\xa4err1\x02\xa4err2\x02\xa4err3\x00",
		`[2, "err1", 2, "err2", 2, "err3", 0]`)

	// one record
	check(
		"\x01\xa4rec1\x00",
		`[1, "rec1", 0]`)

	// a few records
	check(
		"\x01\xa4rec1\x01\xa4rec2\x01\xa4rec3\x00",
		`[1, "rec1", 1, "rec2", 1, "rec3", 0]`)

	// a few records and error
	check(
		"\x02\xa4err1\x01\xa4rec1\x01\xa4rec2\x01\xa4rec3\x00",
		`[2, "err1", 1, "rec1", 1, "rec2", 1, "rec3", 0]`)

	// stat
	check(
		"\x03\xcd\x02+\x00",
		`[3, 555, 0]`)

	// bad cases
	bad("", "EOF")
}
