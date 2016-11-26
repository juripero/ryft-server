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

package rest

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/getryft/ryft-server/rest/codec"
	"github.com/gin-gonic/gin"
)

// Error contains HTTP status and error message.
type Error struct {
	Status  int    `json:"status" msgpack:"status"`                       // HTTP status
	Message string `json:"message,omitempty" msgpack:"message,omitempty"` // error message
	Details string `json:"details,omitempty" msgpack:"details,omitempty"` // error details
}

// NewError creates new server error using status and message.
func NewError(status int, message string) *Error {
	return &Error{
		Status:  status,
		Message: message,
	}
}

// Error get the error as a string.
func (err *Error) Error() string {
	if len(err.Details) != 0 {
		return fmt.Sprintf("%d %s (%s)", err.Status, err.Message, err.Details)
	}

	return fmt.Sprintf("%d %s", err.Status, err.Message)
}

// WithDetails adds additional details to the error.
func (err *Error) WithDetails(details string) *Error {
	err.Details = details
	return err
}

// RecoverFromPanic checks panics and report them via HTTP response.
func RecoverFromPanic(ctx *gin.Context) {
	// check for specific encoder
	reportEncoderError := func(err error) bool {
		// check for specific encoder
		if encI, ok := ctx.Get("encoder"); ok {
			if enc, ok := encI.(codec.Encoder); ok {
				_ = enc.EncodeError(err)
				_ = enc.Close()
				return true
			}
		}

		return false
	}

	// check for panic
	if r := recover(); r != nil {
		var err *Error

		switch v := r.(type) {
		case *Error:
			log.WithError(v).Warnf("Panic recover: server error")
			err = v // report "as is"

		case error:
			log.WithError(v).Warnf("Panic recover: error")
			log.Debugf("stack trace:\n%s", debug.Stack())
			err = NewError(http.StatusInternalServerError, v.Error())

		default:
			log.WithField("error", r).Warnf("Panic recover: object")
			log.Debugf("stack trace:\n%s", debug.Stack())
			err = NewError(http.StatusInternalServerError, fmt.Sprintf("%+v", r))
		}

		// first try to report via encoder...
		if reportEncoderError(err) {
			return
		}

		// report as JSON body
		// err.Message = strings.Replace(err.Message, "\n", " ", -1)
		ctx.IndentedJSON(err.Status, err)
	}
}
