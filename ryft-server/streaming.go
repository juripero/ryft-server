package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/go-fsnotify/fsnotify"
)

func generateJson(records chan IdxRecord, res *os.File, resops chan fsnotify.Op, w io.Writer, dropper chan struct{}) {
	var err error

	w.Write([]byte("["))

	wEncoder := json.NewEncoder(w)
	firstIteration := true
	for r := range records {
		if !firstIteration {
			w.Write([]byte(","))
		}
		r.Data = readDataBlock(res, resops, r.Length)

		log.Printf("writer: writing record... %s, %d", r.File, r.Offset)
		// if err = wEncoder.Encode(r); err != nil {
		// log.Printf("writer: external termination %s, %d sending", r.File, r.Offset)
		// dropper <- struct{}{}
		// log.Printf("writer: external termination %s, %d sent", r.File, r.Offset)
		// return
		// }

		if err = encodeJson(wEncoder, r, 15*time.Second); err != nil {
			log.Printf("writer: external termination %s, %d sending: %s", r.File, r.Offset, err.Error())
			dropper <- struct{}{}
			log.Printf("writer: external termination %s, %d sent", r.File, r.Offset)
			return
		}

		log.Printf("writer: written record %s, %d", r.File, r.Offset)
		firstIteration = false
	}

	w.Write([]byte("]"))
}

func readDataBlock(r io.Reader, resops chan fsnotify.Op, length uint16) (result []byte) {
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

func encodeJson(encoder *json.Encoder, obj interface{}, timeout time.Duration) (err error) {
	ch := make(chan error, 1)
	go func() {
		ch <- encoder.Encode(obj)
	}()

	select {
	case err = <-ch:
		return
	case <-time.After(timeout):
		return fmt.Errorf("Json encoding timeout")
	}
}
