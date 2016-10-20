package main

import (
	"bytes"
	"fmt"

	"github.com/getryft/ryft-server/search"

	json_codec "github.com/getryft/ryft-server/rest/codec/json"
	msgpack_codec "github.com/getryft/ryft-server/rest/codec/msgpack.v2"

	raw_format "github.com/getryft/ryft-server/rest/format/raw"
)

// test JSON simple codec
func testJsonCodec() {
	// empty
	if true {
		b := new(bytes.Buffer)
		c, _ := json_codec.NewSimpleEncoder(b)
		c.Close()
		log("empty: %s", b.String())
	}

	// one record
	if true {
		b := new(bytes.Buffer)
		c, _ := json_codec.NewSimpleEncoder(b)
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test.dat", Offset: 101, Length: 101},
			Data:  []byte("test data")}))
		c.Close()
		log("1-rec: %s", b.String())
	}

	// two records
	if true {
		b := new(bytes.Buffer)
		c, _ := json_codec.NewSimpleEncoder(b)
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test.dat", Offset: 201, Length: 201},
			Data:  []byte("test data")}))
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test2.dat", Offset: 202, Length: 202},
			Data:  []byte("test data N2")}))
		c.Close()
		log("2-rec: %s", b.String())
	}

	// three records
	if true {
		b := new(bytes.Buffer)
		c, _ := json_codec.NewSimpleEncoder(b)
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test.dat", Offset: 301, Length: 301},
			Data:  []byte("test data")}))
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test2.dat", Offset: 302, Length: 302},
			Data:  []byte("test data N2")}))
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test2.dat", Offset: 303, Length: 303},
			Data:  []byte("test data N2")}))
		c.Close()
		log("3-rec: %s", b.String())
	}

	// empty + error
	if true {
		b := new(bytes.Buffer)
		c, _ := json_codec.NewSimpleEncoder(b)
		c.EncodeError(fmt.Errorf("error1"))
		c.Close()
		log("empty: %s", b.String())
	}

	// one record + error
	if true {
		b := new(bytes.Buffer)
		c, _ := json_codec.NewSimpleEncoder(b)
		c.EncodeError(fmt.Errorf("error1"))
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test.dat", Offset: 101, Length: 101},
			Data:  []byte("test data")}))
		c.Close()
		log("1-rec: %s", b.String())
	}

	// two records + error
	if true {
		b := new(bytes.Buffer)
		c, _ := json_codec.NewSimpleEncoder(b)
		c.EncodeError(fmt.Errorf("error1"))
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test.dat", Offset: 201, Length: 201},
			Data:  []byte("test data")}))
		c.EncodeError(fmt.Errorf("error2"))
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test2.dat", Offset: 202, Length: 202},
			Data:  []byte("test data N2")}))
		c.Close()
		log("2-rec: %s", b.String())
	}

	// three records + error
	if true {
		b := new(bytes.Buffer)
		c, _ := json_codec.NewSimpleEncoder(b)
		c.EncodeError(fmt.Errorf("error1"))
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test.dat", Offset: 301, Length: 301},
			Data:  []byte("test data")}))
		c.EncodeError(fmt.Errorf("error2"))
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test2.dat", Offset: 302, Length: 302},
			Data:  []byte("test data N2")}))
		c.EncodeError(fmt.Errorf("error3"))
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test2.dat", Offset: 303, Length: 303},
			Data:  []byte("test data N2")}))
		c.Close()
		log("3-rec: %s", b.String())
	}

	// empty + stat
	if true {
		b := new(bytes.Buffer)
		c, _ := json_codec.NewSimpleEncoder(b)
		c.EncodeStat(raw_format.FromStat(&search.Statistics{Matches: 0, Duration: 1}))
		c.Close()
		log("empty: %s", b.String())
	}

	// one record + stat
	if true {
		b := new(bytes.Buffer)
		c, _ := json_codec.NewSimpleEncoder(b)
		c.EncodeStat(raw_format.FromStat(&search.Statistics{Matches: 1, Duration: 1}))
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test.dat", Offset: 101, Length: 101},
			Data:  []byte("test data")}))
		c.Close()
		log("1-rec: %s", b.String())
	}

	// two records + stat
	if true {
		b := new(bytes.Buffer)
		c, _ := json_codec.NewSimpleEncoder(b)
		c.EncodeStat(raw_format.FromStat(&search.Statistics{Matches: 1, Duration: 1})) // ignored
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test.dat", Offset: 201, Length: 201},
			Data:  []byte("test data")}))
		c.EncodeStat(raw_format.FromStat(&search.Statistics{Matches: 2, Duration: 2}))
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test2.dat", Offset: 202, Length: 202},
			Data:  []byte("test data N2")}))
		c.Close()
		log("2-rec: %s", b.String())
	}

	// three records + stat
	if true {
		b := new(bytes.Buffer)
		c, _ := json_codec.NewSimpleEncoder(b)
		c.EncodeStat(raw_format.FromStat(&search.Statistics{Matches: 1, Duration: 1})) // ignored
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test.dat", Offset: 301, Length: 301},
			Data:  []byte("test data")}))
		c.EncodeStat(raw_format.FromStat(&search.Statistics{Matches: 2, Duration: 2})) // ignored
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test2.dat", Offset: 302, Length: 302},
			Data:  []byte("test data N2")}))
		c.EncodeStat(raw_format.FromStat(&search.Statistics{Matches: 3, Duration: 3}))
		c.EncodeRecord(raw_format.FromRecord(&search.Record{
			Index: search.Index{File: "test2.dat", Offset: 303, Length: 303},
			Data:  []byte("test data N2")}))
		c.Close()
		log("3-rec: %s", b.String())
	}
}

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
