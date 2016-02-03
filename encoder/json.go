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
	"encoding/json"
	"io"
)

// simple JSON encoder
type JsonEncoder struct {
	needSeparator bool
}

func (enc *JsonEncoder) Begin(w io.Writer) error {
	_, err := w.Write([]byte(`{"results":[`))
	return err
}

func (enc *JsonEncoder) End(w io.Writer, errors []error) error {
	return enc.EndWithStats(w, nil, errors)
}

func (enc *JsonEncoder) EndWithStats(w io.Writer, stat interface{}, errors []error) error {
	if _, err := w.Write([]byte(`]`)); err != nil {
		return err
	}
	e := json.NewEncoder(w)

	// errors
	if len(errors) > 0 {
		// convert errors to strings
		messages := make([]string, 0, len(errors))
		for _, e := range errors {
			messages = append(messages, e.Error())
		}

		if _, err := w.Write([]byte(`,"errors":`)); err != nil {
			return err
		}
		if err := e.Encode(messages); err != nil {
			return err
		}
	}

	// statistics
	if stat != nil {
		if _, err := w.Write([]byte(`,"stats":`)); err != nil {
			return err
		}
		if err := e.Encode(stat); err != nil {
			return err
		}
	}

	_, err := w.Write([]byte("}"))
	return err
}

func (enc *JsonEncoder) Write(w io.Writer, item interface{}) error {
	if enc.needSeparator {
		w.Write([]byte(","))
		enc.needSeparator = false
	}
	e := json.NewEncoder(w) // FIXME: do not create encoder each time
	err := e.Encode(item)
	if err == nil {
		enc.needSeparator = true
	}
	return err
}

func (enc *JsonEncoder) WriteStreamError(w io.Writer, err error) bool {
	// JSON fromat doesn't support stream errors
	return false
}