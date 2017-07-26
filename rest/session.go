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

	"gopkg.in/dgrijalva/jwt-go.v2"
)

// Session information
// contains file names and other information
// especially for cluster mode
type Session struct {
	token *jwt.Token
}

// NewSession creates new empty token
func NewSession(alg string) (*Session, error) {
	method := jwt.GetSigningMethod(alg)
	if method == nil {
		return nil, fmt.Errorf("failed to get signing method for %q", alg)
	}

	session := new(Session)
	session.token = jwt.New(method)
	return session, nil // OK
}

// ParseSession parses session from token string
func ParseSession(secret []byte, token string) (*Session, error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse session token: %s", err)
	}

	session := new(Session)
	session.token = t
	return session, nil // OK
}

// Get Session signed string.
func (s *Session) Token(secret []byte) (string, error) {
	return s.token.SignedString(secret)
}

// SetData
func (s *Session) SetData(name string, val interface{}) {
	s.token.Claims[name] = val
}

// GetData
func (s *Session) GetData(name string) interface{} {
	return s.token.Claims[name]
}

// AllData
func (s *Session) AllData() map[string]interface{} {
	return s.token.Claims
}
