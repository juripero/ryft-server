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

	"gopkg.in/vmihailenco/msgpack.v2"
	// "github.com/ugorji/go/codec"
)

// MSGPACK encoder
type MsgPackEncoder struct {
	OmitTags      bool // if we report just data records we can omit tags
	needSeparator bool
}

func (enc *MsgPackEncoder) Begin(w io.Writer) error {
	return nil
}

func (enc *MsgPackEncoder) End(w io.Writer) error {
	if !enc.OmitTags {
		e := msgpack.NewEncoder(w) // FIXME: do not create encoder each time
		_ = e.EncodeUint8(TAG_MsgPackEOF)
	}
	return nil
}

const (
	TAG_MsgPackEOF  uint8 = 0
	TAG_MsgPackItem uint8 = 1
	TAG_MsgPackStat uint8 = 2
)

func (enc *MsgPackEncoder) EndWithStats(w io.Writer, stat interface{}) error {
	//log.Printf("[msgpack]: encode stat: %#v", stat)
	e := msgpack.NewEncoder(w) // FIXME: do not create encoder each time
	if !enc.OmitTags {
		_ = e.EncodeUint8(TAG_MsgPackStat)
	}
	err := e.Encode(stat)
	if !enc.OmitTags {
		_ = e.EncodeUint8(TAG_MsgPackEOF)
	}
	return err
}

func (enc *MsgPackEncoder) Write(w io.Writer, item interface{}) error {
	// log.Printf("[msgpack]: encode item: %#v", item)
	e := msgpack.NewEncoder(w) // FIXME: do not create encoder each time
	if !enc.OmitTags {
		_ = e.EncodeUint8(TAG_MsgPackItem)
	}
	err := e.Encode(item)
	return err
}

// MSGPACK decoder
type MsgPackDecoder struct {
	dec *msgpack.Decoder
}

// NewMsgPackDecoder creates new MSGPACK decoder instance.
func NewMsgPackDecoder(r io.Reader) *MsgPackDecoder {
	return &MsgPackDecoder{dec: msgpack.NewDecoder(r)}
}

func (dec *MsgPackDecoder) NextTag() (uint8, error) {
	return dec.dec.DecodeUint8()
}

// Read decodes next item from the stream.
func (dec *MsgPackDecoder) Next(item interface{}) error {
	return dec.dec.Decode(item)
}
