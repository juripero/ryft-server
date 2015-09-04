package msgpkstream

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/getryft/ryft-rest-api/ryft-server/datapoll"
	"github.com/getryft/ryft-rest-api/ryft-server/records"
	"github.com/ugorji/go/codec"
)

var (
	WriteInterval = time.Second * 20
	mh            codec.MsgpackHandle
)

func Write(source chan records.IdxRecord, res *os.File, w io.Writer, drop chan struct{}) (err error) {

	enc := codec.NewEncoder(w, &mh)

	for r := range source {
		r.Data = datapoll.Next(res, r.Length)
		m := r.OldJsonable()

		if err = encode(enc, m, WriteInterval); err != nil {
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

func encode(enc *codec.Encoder, v interface{}, timeout time.Duration) (err error) {
	ch := make(chan error, 1)
	go func() {
		ch <- enc.Encode(v)
	}()

	select {
	case err = <-ch:
		return
	case <-time.After(timeout):
		return fmt.Errorf("Json encoding timeout")
	}
}
