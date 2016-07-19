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
	"crypto/tls"
	"fmt"

	"gopkg.in/ldap.v2"
)

// LdapAuth contains LDAP related information
type LdapAuth struct {
	Address      string
	QueryFormat  string
	BindUsername string
	BindPassword string
	BaseDN       string

	// TODO: conn *ldap.Conn for caching
}

// NewLDAP returns new LDAP based credentials
func NewLDAP(address, username, password, query, baseDN string) (*LdapAuth, error) {
	a := new(LdapAuth)
	a.Address = address
	a.QueryFormat = query
	a.BindUsername = username
	a.BindPassword = password
	a.BaseDN = baseDN

	return a, nil // OK
}

// reload user credentials
func (a *LdapAuth) Reload() error {
	// TODO: update LDAP options?
	return nil // OK
}

// verify user credentials
func (a *LdapAuth) Verify(username, password string) *UserInfo {
	user, err := a.verify(username, password)
	if err != nil {
		return nil // not found or invalid password
		// or failed to dial LDAP server
	}

	return user
}

// get user's extra data
func (a *LdapAuth) ExtraData(username string) map[string]interface{} {
	return nil // no any data yet
}

// check LDAP server
func (a *LdapAuth) verify(username, password string) (*UserInfo, error) {
	// Connect to LDAP server
	conn, err := ldap.Dial("tcp", a.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to dial LDAP: %s", err)
	}
	defer conn.Close()

	// Reconnect with TLS
	err = conn.StartTLS(&tls.Config{InsecureSkipVerify: true}) // TODO: remove this flag! UNSECURE!!
	if err != nil {
		return nil, fmt.Errorf("failed to use TLS: %s", err)
	}

	// First bind with a read only user
	err = conn.Bind(a.BindUsername, a.BindPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to bind readonly: %s", err)
	}

	// Search for the given username
	req := ldap.NewSearchRequest(a.BaseDN, ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(a.QueryFormat, username),
		[]string{"dn"},
		nil,
	)

	resp, err := conn.Search(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %s", err)
	}

	if len(resp.Entries) != 1 {
		return nil, fmt.Errorf("user does not exist or too many entries returned: %v", req)
	}

	userdn := resp.Entries[0].DN

	// Bind as the user to verify their password
	err = conn.Bind(userdn, password)
	if err != nil {
		return nil, fmt.Errorf("failed to bind user: %s", err)
	}

	user := new(UserInfo)
	user.Name = userdn
	user.Password = password
	user.Home = "/" // TODO: get from LDAP!!!s

	return user, nil // OK
}
