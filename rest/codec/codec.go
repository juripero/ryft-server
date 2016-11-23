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

package codec

import (
	"fmt"
	"io"

	"github.com/getryft/ryft-server/rest/codec/json"
	"github.com/getryft/ryft-server/rest/codec/msgpack.v2"
)

const (
	MIME_JSON     = json.MIME
	MIME_XMSGPACK = msgpack.X_MIME
	MIME_MSGPACK  = msgpack.MIME
)

// Abstract Encoder interface.
type Encoder interface {
	EncodeRecord(rec interface{}) error
	EncodeStat(stat interface{}) error
	EncodeError(err error) error

	io.Closer
}

// Abstract Decoder interface.
type Decoder interface {
	io.Closer
}

// Get list of supported MIME types.
func GetSupportedMimeTypes() []string {
	types := []string{}
	types = append(types, MIME_JSON)
	types = append(types, MIME_MSGPACK)
	types = append(types, MIME_XMSGPACK)
	return types
}

// Create new encoder instance by MIME type.
func NewEncoder(w io.Writer, mime string, stream bool, spark bool) (Encoder, error) {
	switch mime {
	case MIME_JSON:
		if spark {
			return json.NewSparkEncoder(w)
		} else if stream {
			return json.NewStreamEncoder(w)
		} else {
			return json.NewSimpleEncoder(w)
		}
	case MIME_XMSGPACK, MIME_MSGPACK:
		if spark {
			enc, err := msgpack.NewSimpleEncoder(w)
			if err != nil {
				return nil, err
			}
			enc.RecordsOnly = true // Spark format
			return enc, err
		} else if stream {
			return msgpack.NewStreamEncoder(w)
		} else {
			return msgpack.NewSimpleEncoder(w)
		}
	default:
		return nil, fmt.Errorf("%q is unsupported MIME type", mime)
	}
}

// Create new decoder instance by MIME type.
func NewDecoder(r io.Reader, mime string, stream bool) (Decoder, error) {
	return nil, fmt.Errorf("%q not implemented yet", mime)
}
