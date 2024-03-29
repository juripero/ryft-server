/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2018, Ryft Systems, Inc.
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

// LDAP configuration
type LdapConfig struct {
	ServerAddress string `yaml:"server,omitempty"`
	BindUsername  string `yaml:"username,omitempty"`
	BindPassword  string `yaml:"password,omitempty"`
	QueryFormat   string `yaml:"query,omitempty"`
	BaseDN        string `yaml:"basedn,omitempty"`

	InsecureSkipTLS    bool `yaml:"insecure-skip-tls,omitempty"`
	InsecureSkipVerify bool `yaml:"insecure-skip-verify,omitempty"`
}

// LdapAuth contains LDAP related information
type LdapAuth struct {
	LdapConfig

	// TODO: conn *ldap.Conn for caching
}

const (
	// TODO: appropriate attribute names
	attrHomeDir    = "postalAddress"
	attrClusterTag = "postalCode"
)

// NewLDAP returns new LDAP based credentials
func NewLDAP(config LdapConfig) (*LdapAuth, error) {
	a := &LdapAuth{config}
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

// check LDAP server
func (a *LdapAuth) verify(username, password string) (*UserInfo, error) {
	// Connect to LDAP server
	conn, err := ldap.Dial("tcp", a.ServerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to dial LDAP: %s", err)
	}
	defer conn.Close()

	// Reconnect with TLS
	if !a.InsecureSkipTLS {
		err = conn.StartTLS(&tls.Config{InsecureSkipVerify: a.InsecureSkipVerify})
		if err != nil {
			return nil, fmt.Errorf("failed to use TLS: %s", err)
		}
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
		[]string{"dn", attrHomeDir, attrClusterTag},
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
	homeDir := resp.Entries[0].GetAttributeValue(attrHomeDir)
	clusterTag := resp.Entries[0].GetAttributeValue(attrClusterTag)

	// Bind as the user to verify their password
	err = conn.Bind(userdn, password)
	if err != nil {
		return nil, fmt.Errorf("failed to bind user: %s", err)
	}

	user := new(UserInfo)
	user.Name = userdn
	// user.Password = password
	user.HomeDir = homeDir
	user.ClusterTag = clusterTag

	return user, nil // OK
}
