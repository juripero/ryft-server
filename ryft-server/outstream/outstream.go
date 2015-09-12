/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2015, Ryft Systems, Inc.
 * All rights reserved.
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *   this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *   this list of conditions and the following disclaimer in the documentation and/or
 *   other materials provided with the distribution.
 * 3. All advertising materials mentioning features or use of this software must display the following acknowledgement:
 *   This product includes software developed by Ryft Systems, Inc.
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used
 *   to endorse or promote products derived from this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY RYFT SYSTEMS, INC. ''AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL RYFT SYSTEMS, INC. BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 * ============
 */

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
