package msgpack

// TODO: synchronize these tests with v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// test stream encoder
func testStreamEncoder(t *testing.T, populate func(enc *StreamEncoder), expected string) {
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
func testStreamEncoderBad(t *testing.T, populate func(enc *StreamEncoder) error, limit int, expectedError string) {
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

// Test stream encoder
func TestStreamEncoder(t *testing.T) {
	// empty
	testStreamEncoder(t, func(enc *StreamEncoder) {
		// do nothing
	}, "\x00")

	// one error
	testStreamEncoder(t, func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(nil)) // ignored
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
	}, "\x02\xa4err1\x00")

	// a few errors
	testStreamEncoder(t, func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err2")))
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err3")))
	}, "\x02\xa4err1\x02\xa4err2\x02\xa4err3\x00")

	// one record
	testStreamEncoder(t, func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeRecord(nil)) // ignored
		assert.NoError(t, enc.EncodeRecord("rec1"))
	}, "\x01\xa4rec1\x00")

	// a few records
	testStreamEncoder(t, func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeRecord("rec1"))
		assert.NoError(t, enc.EncodeRecord("rec2"))
		assert.NoError(t, enc.EncodeRecord("rec3"))
	}, "\x01\xa4rec1\x01\xa4rec2\x01\xa4rec3\x00")

	// a few records and error
	testStreamEncoder(t, func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeError(fmt.Errorf("err1")))
		assert.NoError(t, enc.EncodeRecord("rec1"))
		assert.NoError(t, enc.EncodeRecord("rec2"))
		assert.NoError(t, enc.EncodeRecord("rec3"))
	}, "\x02\xa4err1\x01\xa4rec1\x01\xa4rec2\x01\xa4rec3\x00")

	// stat
	testStreamEncoder(t, func(enc *StreamEncoder) {
		assert.NoError(t, enc.EncodeStat(nil)) // ignored
		assert.NoError(t, enc.EncodeStat(555))
	}, "\x03\xcd\x02+\x00")

	// bad cases
	testStreamEncoderBad(t, func(enc *StreamEncoder) error {
		if err := enc.EncodeRecord("rec1"); err != nil {
			return err
		}
		return nil
	}, 0, "EOF") // EncodeRecord (tag)
	testStreamEncoderBad(t, func(enc *StreamEncoder) error {
		if err := enc.EncodeRecord("rec1"); err != nil {
			return err
		}
		return nil
	}, 4, "EOF") // EncodeRecord (record itself)
	testStreamEncoderBad(t, func(enc *StreamEncoder) error {
		if err := enc.EncodeStat("stat1"); err != nil {
			return err
		}
		return nil
	}, 0, "EOF") // EncodeStat (tag)
	testStreamEncoderBad(t, func(enc *StreamEncoder) error {
		if err := enc.EncodeStat("stat1"); err != nil {
			return err
		}
		return nil
	}, 4, "EOF") // EncodeStat (record itself)
	testStreamEncoderBad(t, func(enc *StreamEncoder) error {
		if err := enc.EncodeError(fmt.Errorf("err1")); err != nil {
			return err
		}
		return nil
	}, 0, "EOF") // EncodeStat (tag)
	testStreamEncoderBad(t, func(enc *StreamEncoder) error {
		if err := enc.EncodeError(fmt.Errorf("err1")); err != nil {
			return err
		}
		return nil
	}, 4, "EOF") // EncodeStat (record itself)
	testStreamEncoderBad(t, func(enc *StreamEncoder) error {
		return nil
	}, 0, "EOF") // Close
}

// test stream JSON decoder
func testStreamDecoder(t *testing.T, data string, expected string) {
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
func testStreamDecoderBad(t *testing.T, data string, expectedError string) {
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

// Test stream decoder
func TestStreamDecoder(t *testing.T) {
	// empty
	testStreamDecoder(t,
		"\x00",
		`[0]`)

	// one error
	testStreamDecoder(t,
		"\x02\xa4err1\x00",
		`[2, "err1", 0]`)

	// a few errors
	testStreamDecoder(t,
		"\x02\xa4err1\x02\xa4err2\x02\xa4err3\x00",
		`[2, "err1", 2, "err2", 2, "err3", 0]`)

	// one record
	testStreamDecoder(t,
		"\x01\xa4rec1\x00",
		`[1, "rec1", 0]`)

	// a few records
	testStreamDecoder(t,
		"\x01\xa4rec1\x01\xa4rec2\x01\xa4rec3\x00",
		`[1, "rec1", 1, "rec2", 1, "rec3", 0]`)

	// a few records and error
	testStreamDecoder(t,
		"\x02\xa4err1\x01\xa4rec1\x01\xa4rec2\x01\xa4rec3\x00",
		`[2, "err1", 1, "rec1", 1, "rec2", 1, "rec3", 0]`)

	// stat
	testStreamDecoder(t,
		"\x03\xcd\x02+\x00",
		`[3, 555, 0]`)

	// bad cases
	testStreamDecoderBad(t, "", "EOF")
}
