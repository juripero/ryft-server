package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	"github.com/go-fsnotify/fsnotify"
)

func generateJson(records chan IdxRecord, res *os.File, resops chan fsnotify.Op, w io.Writer) {
	var err error

	w.Write([]byte("["))

	wEncoder := json.NewEncoder(w)
	firstIteration := true
	for r := range records {
		if !firstIteration {
			w.Write([]byte(","))
		}

		log.Printf("generate: processing offset=%d...", r.Offset)
		r.Data = readDataBlock(res, r.Length)
		log.Printf("generate: processed offset=%d, len=%d", r.Offset, len(r.Data))

		if err = wEncoder.Encode(r); err != nil {
			log.Printf("Encoding error: %s", err.Error())
			return
		}
		firstIteration = false
	}

	w.Write([]byte("]"))
}

func readDataBlock(r io.Reader, length uint16) (result []byte) {
	var total uint16 = 0
	for total < length {
		data := make([]byte, length-total)
		n, _ := r.Read(data)
		if n != 0 {
			result = append(result, data...)
			total = total + uint16(n)
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}
	return
}
