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

package encoder

import (
	"io"
	"fmt"
	"time"
	"encoding/json"
)

type JsonEncoder struct {
	Encoder
	needSeparator bool
}




func (enc *JsonEncoder) Begin(w io.Writer) error {
	_, err := w.Write([]byte("["))
	return err
}

func (enc *JsonEncoder) End(w io.Writer) error {
	_, err := w.Write([]byte("]"))
	return err
}

func (enc *JsonEncoder) Write(w io.Writer, itm interface{}) error {
	if enc.needSeparator {
		w.Write([]byte(","))
		enc.needSeparator = false
	}
	wEncoder := json.NewEncoder(w)
	err := jsonEncode(wEncoder, itm, WriteInterval)
	if err == nil {
		enc.needSeparator = true
	}
	return err
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

