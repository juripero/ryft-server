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
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/appleboy/gin-jwt.v1"
)

const (
	tokenAttrRoles   = "roles"
	tokenAttrHomeDir = "home-dir"
	tokenAttrCluster = "cluster-tag"

	AdminRole = "admin"
)

// UserInfo is a user credentials and related information such a home directory.
type UserInfo struct {
	Name       string   `json:"username" yaml:"username"`
	Password   string   `json:"password,omitempty" yaml:"password,omitempty"` // for backward compatibility, should be empty
	Passhash   string   `json:"passhash,omitempty" yaml:"passhash,omitempty"` // hashed password
	Roles      []string `json:"roles,omitempty" yaml:"roles,omitempty"`       // "admin", "user"
	HomeDir    string   `json:"home,omitempty" yaml:"home,omitempty"`
	ClusterTag string   `json:"cluster-tag,omitempty" yaml:"cluster-tag,omitempty"`
}

// WipeOut creates copy of user with no sensitive information
func (u *UserInfo) WipeOut() *UserInfo {
	w := *u // copy
	w.Password = ""
	w.Passhash = ""
	return &w
}

// HasRole checks if user has requested role
func (u *UserInfo) HasRole(role string) bool {
	for _, r := range u.Roles {
		if role == r {
			return true
		}
	}

	return false // not found
}

// used to verify user's password
type Provider interface {
	Reload() error
	Verify(username string, password string) *UserInfo
}

// used to manage set of users
type Manager interface {
	// get the list of all users.
	GetAllUsers() ([]*UserInfo, error)

	// get the list of specified users
	GetUsers(names []string) ([]*UserInfo, error)

	// create new user
	CreateNew(user *UserInfo) (*UserInfo, error)
}

type Middleware struct {
	provider Provider
	realm    string
	jwt      *jwt.GinJWTMiddleware

	userCache     map[string]*UserInfo
	userCacheLock sync.Mutex
}

// Create new middleware with custom auth provider
func NewMiddleware(provider Provider, realm string) *Middleware {
	if len(realm) == 0 {
		realm = "Authorization Required"
	}

	mw := new(Middleware)
	mw.provider = provider
	mw.realm = realm
	mw.userCache = make(map[string]*UserInfo)

	return mw
}

// Enable JWT authentication
// `alg` - signing algorithm
func (mw *Middleware) EnableJwt(key []byte, alg string, tokenLifetime time.Duration) {
	mw.jwt = new(jwt.GinJWTMiddleware)
	mw.jwt.SigningAlgorithm = alg
	mw.jwt.PayloadFunc = mw.payload
	mw.jwt.Realm = mw.realm
	mw.jwt.Key = key
	mw.jwt.Timeout = tokenLifetime
	mw.jwt.MaxRefresh = 10 * tokenLifetime
	mw.jwt.Authenticator = mw.authenticator
	mw.jwt.Authorizator = mw.authorizator
	mw.jwt.Unauthorized = mw.unauthorized
}

// Login handler for JWT
func (mw *Middleware) LoginHandler() gin.HandlerFunc {
	return mw.jwt.LoginHandler
}

// Refresh token handler for JWT
func (mw *Middleware) RefreshHandler() gin.HandlerFunc {
	return mw.jwt.RefreshHandler
}

// Authentication middleware function
// tries Basic Auth first, then JWT
func (mw *Middleware) Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Search user in the slice of allowed credentials
		h := c.Request.Header.Get("Authorization")

		username, password, ok, err := parseBasicAuth(h)
		if ok && err == nil { // basic authentication
			user := mw.provider.Verify(username, password)
			if user == nil {
				// Credentials doesn't match, we return 401 and abort handlers chain.
				c.Header("WWW-Authenticate", "Basic realm="+strconv.Quote(mw.realm))
				c.AbortWithStatus(http.StatusUnauthorized)
			} else {
				// The user credentials was found!
				// pass user info as key "user" in this context
				// the user can be read later using c.MustGet(gin.AuthUserKey)
				mw.putUserToCache(username, user)
				c.Set(gin.AuthUserKey, user)
			}
		} else if mw.jwt != nil && len(h) != 0 {
			f := mw.jwt.MiddlewareFunc()
			f(c) // JWT work
		} else {
			c.Header("WWW-Authenticate", "Basic realm="+strconv.Quote(mw.realm))
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

// authenticator: checks userId exists and password is correct
func (mw *Middleware) authenticator(userId string, password string, c *gin.Context) (string, bool) {
	user := mw.provider.Verify(userId, password)
	if user != nil {
		mw.putUserToCache(userId, user)
		c.Set(gin.AuthUserKey, user)
	}
	return userId, user != nil
}

// authorizator: all logged in users have access
func (mw *Middleware) authorizator(userId string, ctx *gin.Context) bool {
	if user := getUserFromJwt(userId, ctx); user != nil {
		ctx.Set(gin.AuthUserKey, user)
	}
	return true
}

// creates user info based on token attributes
func getUserFromJwt(userId string, ctx *gin.Context) *UserInfo {
	if ival, ok := ctx.Get("JWT_PAYLOAD"); ok && ival != nil {
		if val, ok := ival.(map[string]interface{}); ok {
			user := new(UserInfo)
			user.Name, _ = val["id"].(string) // userId
			user.Roles, _ = val[tokenAttrRoles].([]string)
			user.HomeDir, _ = val[tokenAttrHomeDir].(string)
			user.ClusterTag, _ = val[tokenAttrCluster].(string)
			return user
		}
	}

	return nil
}

// put user to cache
func (mw *Middleware) putUserToCache(username string, user *UserInfo) {
	mw.userCacheLock.Lock()
	defer mw.userCacheLock.Unlock()

	// put to cache under lock
	mw.userCache[username] = user
}

// report unauthorized access
func (mw *Middleware) unauthorized(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"code":    code,
		"message": message,
	})
}

// get additional payload
func (mw *Middleware) payload(userId string) map[string]interface{} {
	mw.userCacheLock.Lock()
	defer mw.userCacheLock.Unlock()

	if user, ok := mw.userCache[userId]; ok {
		return map[string]interface{}{
			tokenAttrRoles:   user.Roles,
			tokenAttrHomeDir: user.HomeDir,
			tokenAttrCluster: user.ClusterTag,
		}
	}

	return nil // no data yet
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
