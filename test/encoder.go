package main

import (
	"bytes"
	"fmt"

	"github.com/getryft/ryft-server/codec/json"
	"github.com/getryft/ryft-server/format/raw"
	"github.com/getryft/ryft-server/search"
)

func testEncoder() {
	b1 := new(bytes.Buffer)
	e1, _ := json.NewSimpleEncoder(b1)

	b2 := new(bytes.Buffer)
	e2, _ := json.NewStreamEncoder(b2)

	r1 := new(search.Record)
	r1.Index.File = "test1.txt"
	r1.Index.Offset = 100
	r1.Index.Length = 200
	r1.Data = []byte("record1")
	rx1 := raw.FromRecord(r1)
	e1.EncodeRecord(rx1)
	e2.EncodeRecord(rx1)

	e1.EncodeError(fmt.Errorf("error 1"))
	e1.EncodeError(fmt.Errorf("error 2"))
	e2.EncodeError(fmt.Errorf("error 1*"))
	e2.EncodeError(fmt.Errorf("error 2*"))

	r2 := new(search.Record)
	r2.Index.File = "test2.txt"
	r2.Index.Offset = 300
	r2.Index.Length = 400
	r2.Data = []byte("record2")
	rx2 := raw.FromRecord(r2)
	e1.EncodeRecord(rx2)
	e2.EncodeRecord(rx2)

	s1 := search.NewStat()
	s1.Matches = 100
	s1.Duration = 200
	s1.TotalBytes = 300
	sx1 := raw.FromStat(s1)
	e1.EncodeStat(sx1)
	e2.EncodeStat(sx1)

	e1.Close()
	e2.Close()

	log("simple:\n%s\n", b1.String())
	log("stream:\n%s\n", b2.String())
}
