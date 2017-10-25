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

	"github.com/getryft/ryft-server/middleware/auth"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// UserParams contains all the bound parameters for the /user endpoint.
type UserParams struct {
	Names []string `form:"name" json:"names,omitempty" msgpack:"names,omitempty"`
}

// Handle GET /user endpoint get all users
func (server *Server) DoUserGet(ctx *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(ctx)

	var user *auth.UserInfo
	if user_, ok := ctx.Get(gin.AuthUserKey); !ok {
		panic(NewError(http.StatusUnauthorized, "no authenticated user found"))
	} else if user, ok = user_.(*auth.UserInfo); !ok {
		panic(NewError(http.StatusInternalServerError, "no authenticated user found"))
	}

	// parse request parameters
	var params UserParams
	/*if err := bindOptionalJson(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request JSON parameters"))
	}*/
	if err := binding.Form.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	var res []*auth.UserInfo
	var err error
	if len(params.Names) == 0 {
		if user.HasRole(auth.AdminRole) {
			res, err = server.AuthManager.GetAllUsers()
		} else {
			res = append(res, user.WipeOut())
		}
	} else {
		if user.HasRole(auth.AdminRole) {
			res, err = server.AuthManager.GetUsers(params.Names)
		} else {
			for _, name := range params.Names {
				if user.Name == name {
					res = append(res, user.WipeOut())
				} else {
					panic(NewError(http.StatusForbidden,
						fmt.Sprintf(`access to "%s" denied`, name)))
				}
			}
		}
	}

	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get users"))
	}

	ctx.JSON(http.StatusOK, res)
}

// Handle POST /user endpoint - create new user
func (server *Server) DoUserPost(ctx *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(ctx)

	var user *auth.UserInfo
	if user_, ok := ctx.Get(gin.AuthUserKey); !ok {
		panic(NewError(http.StatusUnauthorized, "no authenticated user found"))
	} else if user, ok = user_.(*auth.UserInfo); !ok {
		panic(NewError(http.StatusInternalServerError, "no authenticated user found"))
	}

	// parse request parameters
	var newUser auth.UserInfo
	if err := bindOptionalJson(ctx.Request, &newUser); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request JSON parameters"))
	}

	// check required parameters
	if len(newUser.Name) == 0 {
		panic(NewError(http.StatusBadRequest, "no username provided"))
	}
	if len(newUser.Password) == 0 {
		panic(NewError(http.StatusBadRequest, "no password provided"))
	}

	if !user.HasRole(auth.AdminRole) {
		panic(NewError(http.StatusForbidden, "only admin can create new users"))
	}

	log.Debugf("[%s/auth]: creating new user...", CORE)
	res, err := server.AuthManager.CreateNew(&newUser)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to create new user"))
	}

	ctx.JSON(http.StatusOK, res)
}
