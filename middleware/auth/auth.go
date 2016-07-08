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

package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserInfo is a user credentials and related information such a home directory.
type UserInfo struct {
	Name     string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Home     string `json:"home" yaml:"home"`
}

type Provider interface {
	Verify(username string, password string) *UserInfo
}

func MiddlewareFunc(provider Provider, realm string) gin.HandlerFunc {
	if len(realm) == 0 {
		realm = "Authorization Required"
	}
	realm = "Basic realm=" + strconv.Quote(realm)

	return func(c *gin.Context) {
		// Search user in the slice of allowed credentials
		h := c.Request.Header.Get("Authorization")

		username, password, ok, err := parseBasicAuth(h)
		fmt.Printf("auth: h:%q %q:%q %v %v\n", h, username, password, ok, err)
		if ok && err == nil { // basic
			// Search user in the slice of allowed credentials
			fmt.Printf("users: %+v\n", provider)
			user := provider.Verify(username, password)
			if user == nil {
				// Credentials doesn't match, we return 401 and abort handlers chain.
				c.Header("WWW-Authenticate", realm)
				c.AbortWithStatus(http.StatusUnauthorized)
				fmt.Printf("reported 401\n")
			} else {
				// The user credentials was found!
				// pass user info as key "user" in this context
				// the user can be read later using c.MustGet(gin.AuthUserKey)
				c.Set(gin.AuthUserKey, user)
				fmt.Printf("authenticated to %v\n", user)
			}
		} else {
			// TODO: JWT
			c.Header("WWW-Authenticate", realm)
			c.AbortWithStatus(http.StatusUnauthorized)
			fmt.Printf("JWT: reported 401\n")
		}
	}
}

// Try to decode Authorization header (basic) to get username and password
// ok=true means it's basic authentication
func parseBasicAuth(auth string) (username, password string, ok bool, err error) {
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return
	}

	ok = true // it's Basic Auth

	// decode header
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		err = fmt.Errorf("invalid format")
		return
	}

	return cs[:s], cs[s+1:], ok, nil
}
