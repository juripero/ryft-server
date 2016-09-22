package main

import (
	"bytes"
	// "encoding/hex"
	"encoding/base64"
	"fmt"

	"github.com/getryft/ryft-server/codec"
	"github.com/getryft/ryft-server/codec/json"
	"github.com/getryft/ryft-server/codec/msgpack.v2"
	"github.com/getryft/ryft-server/format/raw"
	"github.com/getryft/ryft-server/search"
)

func testEncoder() {
	b1 := new(bytes.Buffer)
	e1, _ := json.NewSimpleEncoder(b1)

	b2 := new(bytes.Buffer)
	e2, _ := json.NewStreamEncoder(b2)

	b3 := new(bytes.Buffer)
	e3, _ := msgpack.NewSimpleEncoder(b3)

	b4 := new(bytes.Buffer)
	e4, _ := msgpack.NewStreamEncoder(b4)

	encs := []codec.Encoder{e1, e2, e3, e4}

	r1 := new(search.Record)
	r1.Index.File = "test1.txt"
	r1.Index.Offset = 100
	r1.Index.Length = 200
	r1.Data = []byte("record1")
	rx1 := raw.FromRecord(r1)
	for _, e := range encs {
		e.EncodeRecord(rx1)
	}

	err1 := fmt.Errorf("error-1")
	err2 := fmt.Errorf("error-2")
	for _, e := range encs {
		e.EncodeError(err1)
		e.EncodeError(err2)
	}

	r2 := new(search.Record)
	r2.Index.File = "test2.txt"
	r2.Index.Offset = 300
	r2.Index.Length = 400
	r2.Data = []byte("record2")
	rx2 := raw.FromRecord(r2)
	for _, e := range encs {
		e.EncodeRecord(rx2)
	}

	s1 := search.NewStat("")
	s1.Matches = 100
	s1.Duration = 200
	s1.TotalBytes = 300
	sx1 := raw.FromStat(s1)
	for _, e := range encs {
		e.EncodeStat(sx1)
	}

	for _, e := range encs {
		e.Close()
	}

	log("json simple:\n%s\n", b1.String())
	log("json stream:\n%s\n", b2.String())
	log("msgpack simple:\n%s\n", bytesToStr(b3.Bytes()))
	log("msgpack stream:\n%s\n", bytesToStr(b4.Bytes()))
}

// convert bytes to string
func bytesToStr(b []byte) string {
	w := new(bytes.Buffer)

	e := base64.NewEncoder(base64.StdEncoding, w)
	e.Write(b)
	e.Close()

	return w.String()
}
