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

	//"gopkg.in/vmihailenco/msgpack.v2"
	"github.com/ugorji/go/codec"
)

// MSGPACK encoder
type MsgPackEncoder struct {
	OmitTags      bool // if we report just data records we can omit tags
	needSeparator bool

	handle *codec.MsgpackHandle
}

const (
	TAG_MsgPackEOF uint8 = iota
	TAG_MsgPackItem
	TAG_MsgPackStat
	TAG_MsgPackError
)

func NewMsgPackEncoder() *MsgPackEncoder {
	enc := new(MsgPackEncoder)
	enc.handle = new(codec.MsgpackHandle)
	return enc
}

func (enc *MsgPackEncoder) Begin(w io.Writer) error {
	return nil
}

func (enc *MsgPackEncoder) End(w io.Writer, errors []error) error {
	return enc.EndWithStats(w, nil, errors)
}

func (enc *MsgPackEncoder) EndWithStats(w io.Writer, stat interface{}, errors []error) error {
	//log.Printf("[msgpack]: encode stat: %#v", stat)
	e := codec.NewEncoder(w, enc.handle) // FIXME: do not create encoder each time

	if len(errors) > 0 {
		for _, err := range errors {
			if !enc.OmitTags {
				_ = e.Encode(TAG_MsgPackError)
			}
			_ = e.Encode(err.Error())
		}
	}

	if stat != nil {
		if !enc.OmitTags {
			_ = e.Encode(TAG_MsgPackStat)
		}
		_ = e.Encode(stat)
	}

	if !enc.OmitTags {
		_ = e.Encode(TAG_MsgPackEOF)
	}
	return nil
}

func (enc *MsgPackEncoder) Write(w io.Writer, item interface{}) error {
	// log.Printf("[msgpack]: encode item: %#v", item)
	e := codec.NewEncoder(w, enc.handle) // FIXME: do not create encoder each time
	if !enc.OmitTags {
		_ = e.Encode(TAG_MsgPackItem)
	}
	err := e.Encode(item)
	return err
}

func (enc *MsgPackEncoder) WriteStreamError(w io.Writer, err error) bool {
	if !enc.OmitTags {
		e := codec.NewEncoder(w, enc.handle) // FIXME: do not create encoder each time
		_ = e.Encode(TAG_MsgPackError)
		_ = e.Encode(err.Error())
		return true
	}
	return false
}

// MSGPACK decoder
type MsgPackDecoder struct {
	dec *codec.Decoder
}

// NewMsgPackDecoder creates new MSGPACK decoder instance.
func NewMsgPackDecoder(r io.Reader) *MsgPackDecoder {
	handle := new(codec.MsgpackHandle)
	return &MsgPackDecoder{dec: codec.NewDecoder(r, handle)}
}

func (dec *MsgPackDecoder) NextTag() (uint8, error) {
	var tag uint8
	err := dec.dec.Decode(&tag)
	return tag, err
}

// Read decodes next item from the stream.
func (dec *MsgPackDecoder) Next(item interface{}) error {
	return dec.dec.Decode(item)
}
