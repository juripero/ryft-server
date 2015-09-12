package outstream

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/getryft/ryft-rest-api/ryft-server/binding"
	"github.com/getryft/ryft-rest-api/ryft-server/datapoll"
	"github.com/getryft/ryft-rest-api/ryft-server/records"
	"github.com/ugorji/go/codec"
)

var WriteInterval = time.Second * 20

func Write(s *binding.Search, source chan records.IdxRecord, res *os.File, w io.Writer, drop chan struct{}) (err error) {
	if s.IsOutJson() {
		w.Write([]byte("["))
		wEncoder := json.NewEncoder(w)
		firstIteration := true
		for r := range source {

			r.Data = datapoll.Next(res, r.Length)

			obj, recerr := s.FormatConvertor(r)
			if recerr != nil {
				log.Printf("%s: DATA RECORD OFFSET=%d CAN NOT BE CONVERTED WITH ERROR: %s", res.Name(), r.Offset, recerr.Error())
				if r.Data != nil {
					log.Printf("%s:!DATA RECORD OFFSET=%d: `%s`", res.Name(), r.Offset, string(r.Data))
				}
				continue
			}

			if !firstIteration {
				w.Write([]byte(","))
			}


			if err = jsonEncode(wEncoder, obj, WriteInterval); err != nil {
				log.Printf("%s: DATA ENCODED OFFSET=%d WITH ERROR: %s", res.Name(), r.Offset, err.Error())
				drop <- struct{}{}

				for range source {
				}
				log.Printf("%s: DROPPED CONNECTION", res.Name())
				return
			}

			firstIteration = false
		}
		w.Write([]byte("]"))
		return
	}

	if s.IsOutMsgpk() {
		var mh codec.MsgpackHandle
		enc := codec.NewEncoder(w, &mh)

		for r := range source {
			r.Data = datapoll.Next(res, r.Length)
			obj, recerr := s.FormatConvertor(r)
			if recerr != nil {
				log.Printf("%s: DATA RECORD OFFSET=%d CAN NOT BE CONVERTED WITH ERROR: %s", res.Name(), r.Offset, recerr.Error())
				if r.Data != nil {
					log.Printf("%s:!DATA RECORD OFFSET=%d: `%s`", res.Name(), r.Offset, string(r.Data))
				}

				continue
			}

			if err = msgpkEncode(enc, obj, WriteInterval); err != nil {
				log.Printf("%s: DATA ENCODED OFFSET=%d WITH ERROR: %s", res.Name(), r.Offset, err.Error())
				drop <- struct{}{}

				for range source {
				}
				log.Printf("%s: DROPPED CONNECTION", res.Name())
				return
			}
		}
		return
	}

	return
}

func jsonEncode(enc *json.Encoder, obj interface{}, timeout time.Duration) (err error) {
	ch := make(chan error, 1)
	go func() {
		ch <- enc.Encode(obj)
	}()

	select {
	case err = <-ch:
		return
	case <-time.After(timeout):
		return fmt.Errorf("Json encoding timeout")
	}
}

func msgpkEncode(enc *codec.Encoder, v interface{}, timeout time.Duration) (err error) {
	ch := make(chan error, 1)
	go func() {
		ch <- enc.Encode(v)
	}()

	select {
	case err = <-ch:
		return
	case <-time.After(timeout):
		return fmt.Errorf("Msgpk encoding timeout")
	}
}
