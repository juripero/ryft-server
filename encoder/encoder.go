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
	"net/http"
	"time"

	"github.com/getryft/ryft-server/srverr"
	"github.com/gin-gonic/gin"
)

const (
	MIMEJSON     = "application/json"
	MIMEMSGPACKX = "application/x-msgpack"
	MIMEMSGPACK  = "application/msgpack"

	WriteInterval = time.Second * 20
	CTXKEY        = "encoder-detected"
)

type Encoder interface {
	Begin(w io.Writer) error
	End(w io.Writer) error
	EndWithStats(w io.Writer, stats map[string]interface{}) error
	Write(w io.Writer, itm interface{}) error
}

func GetSupportedMimeTypes() []string {
	return []string{MIMEJSON, MIMEMSGPACK, MIMEMSGPACKX}
}

func GetByMimeType(mime string) (Encoder, error) {
	switch mime {
	case MIMEJSON:
		return new(JsonEncoder), nil
	case MIMEMSGPACKX, MIMEMSGPACK:
		return new(MsgPackEncoder), nil
	default:
		return nil, fmt.Errorf("Unsupported mime type: %s", mime)
	}
}

func Detect(c *gin.Context) {
	accept := c.NegotiateFormat(GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = MIMEJSON
	}
	c.Header("Content-Type", accept)

	// setting up encoder to respond with requested format
	if enc, err := GetByMimeType(accept); err != nil {
		panic(srverr.New(http.StatusBadRequest, err.Error()))
	} else {
		c.Set(CTXKEY, enc)
	}
}

func FromContext(c *gin.Context) Encoder {
	// TODO add handlers for null value and report 400 error
	return c.MustGet(CTXKEY).(Encoder)
}
