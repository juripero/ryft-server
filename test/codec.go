package main

import (
	"bytes"
	"fmt"

	"github.com/getryft/ryft-server/search"

	json_codec "github.com/getryft/ryft-server/rest/codec/json"
	msgpack_codec "github.com/getryft/ryft-server/rest/codec/msgpack.v2"

	raw_format "github.com/getryft/ryft-server/rest/format/raw"
)

// test msgpack codec and raw format
func testMsgpackFormat() {
	idx := search.Index{}
	idx.File = "test.file.txt"
	idx.Offset = 12345
	idx.Length = 123
	idx.Fuzziness = 100
	idx.Host = "localhost"
	rec := new(search.Record)
	rec.Index = idx
	rec.Data = []byte("test data")

	b := new(bytes.Buffer)
	enc, _ := msgpack_codec.NewStreamEncoder(b)
	enc.EncodeRecord(raw_format.FromRecord(rec))
	enc.Close()

	log("%s", b.String())
}
