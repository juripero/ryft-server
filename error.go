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

package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ServerError struct {
	Status  int
	Message string
	Details string
}

func (err *ServerError) Error() string {
	return fmt.Sprintf("%d %s", err.Status, err.Message)
}

func NewServerError(status int, message string) *ServerError {
	return &ServerError{status, message, ""}
}

func NewServerErrorWithDetails(status int, message string, details string) *ServerError {
	return &ServerError{status, message, details}
}

func RecoverFromPanic(c *gin.Context) {
	if r := recover(); r != nil {
		if err, ok := r.(*ServerError); ok {
			log.WithField("status", err.Status).WithError(err).Warnf("Panic recovered server error")
			if len(err.Details) > 0 {
				c.IndentedJSON(err.Status, gin.H{"message": fmt.Sprintf("%s", strings.Replace(err.Message, "\n", " ", -1)), "status": err.Status, "details": err.Details})
			} else {
				c.IndentedJSON(err.Status, gin.H{"message": fmt.Sprintf("%s", strings.Replace(err.Message, "\n", " ", -1)), "status": err.Status})
			}

			return
		}

		if err, ok := r.(error); ok {
			log.WithError(err).Warnf("Panic recovered unknown error")
			//			debug.PrintStack()
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("%v", err.Error()), "status": http.StatusInternalServerError})
			return
		}

		log.WithField("error", r).Warnf("Panic recovered with object")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("%+v", r), "status": http.StatusInternalServerError})
	}
}
