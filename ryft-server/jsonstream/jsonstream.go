package jsonstream

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/DataArt/ryft-rest-api/ryft-server/datapoll"
	"github.com/DataArt/ryft-rest-api/ryft-server/records"
)

var WriteInterval = time.Second * 20

func Write(source chan records.IdxRecord, res *os.File, w io.Writer, drop chan struct{}) (err error) {
	w.Write([]byte("["))
	wEncoder := json.NewEncoder(w)
	firstIteration := true
	for r := range source {
		log.Printf("%s: RECV OFFSET=%d", res.Name(), r.Offset)
		if !firstIteration {
			w.Write([]byte(","))
		}

		log.Printf("%s: DATA READING OFFSET=%d...", res.Name(), r.Offset)
		r.Data = datapoll.Next(res, r.Length)
		log.Printf("%s: DATA READ COMPLETE OFFSET=%d, DATA ENCODING...", res.Name(), r.Offset)
		if err = encode(wEncoder, r, WriteInterval); err != nil {
			log.Printf("%s: DATA ENCODED OFFSET=%d WITH ERROR: %s", res.Name(), r.Offset, err.Error())
			drop <- struct{}{}

			for range source {
			}
			log.Printf("%s: DROPPED CONNECTION", res.Name())
			return
		}
		log.Printf("%s: DATA ENCODED OFFSET=%d", res.Name(), r.Offset)

		firstIteration = false
	}
	w.Write([]byte("]"))
	return
}

func encode(encoder *json.Encoder, obj interface{}, timeout time.Duration) (err error) {
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
