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
	"fmt"
	"io"
)

const (
	MIME_JSON     = "application/json"
	MIME_XMSGPACK = "application/x-msgpack"
	MIME_MSGPACK  = "application/msgpack"
)

// abstract Encoder interface
type Encoder interface {
	Begin(w io.Writer) error
	End(w io.Writer, errors []error) error
	EndWithStats(w io.Writer, stat interface{}, errors []error) error
	Write(w io.Writer, itm interface{}) error

	// if stream errors are not supported, return `false`
	WriteStreamError(w io.Writer, err error) bool
}

// get list of supported MIME types
func GetSupportedMimeTypes() []string {
	types := []string{}
	types = append(types, MIME_JSON)
	types = append(types, MIME_MSGPACK)
	types = append(types, MIME_XMSGPACK)
	return types
}

// get encoder instance by MIME type
func GetByMimeType(mime string) (Encoder, error) {
	switch mime {
	case MIME_JSON:
		return new(JsonEncoder), nil
	case MIME_XMSGPACK, MIME_MSGPACK:
		return new(MsgPackEncoder), nil
	default:
		return nil, fmt.Errorf("Unsupported mime type: %s", mime)
	}
}
