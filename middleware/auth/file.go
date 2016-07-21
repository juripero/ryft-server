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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// FileAuth contains dictionary of users
type FileAuth struct {
	Users    map[string]*UserInfo
	FileName string
}

// NewFile returns new File based credentials
func NewFile(fileName string) (*FileAuth, error) {
	users, err := readUsersFile(fileName)
	if err != nil {
		return nil, err
	}

	// check for duplicates
	unique, err := checkForDuplicates(users)
	if err != nil {
		return nil, err
	}

	f := new(FileAuth)
	f.Users = unique
	f.FileName = fileName
	return f, nil // OK
}

// reload user credentials
func (f *FileAuth) Reload() error {
	users, err := readUsersFile(f.FileName)
	if err != nil {
		return err
	}

	// check for duplicates
	unique, err := checkForDuplicates(users)
	if err != nil {
		return err
	}

	f.Users = unique
	return nil // OK
}

// find user credentials
func (f *FileAuth) FindUser(username string) *UserInfo {
	if u, ok := f.Users[username]; ok {
		return u // found
	}

	return nil // not found
}

// verify user credentials
func (f *FileAuth) Verify(username, password string) *UserInfo {
	if u, ok := f.Users[username]; ok {
		if u.Password == password {
			return u // verified!
		}
	}

	return nil // not found or invalid password
}

// get user's extra data
func (f *FileAuth) ExtraData(username string) map[string]interface{} {
	if u, ok := f.Users[username]; ok {
		return map[string]interface{}{
			"home-dir":    u.Home,
			"cluster-tag": u.ClusterTag,
		}
	}

	return nil // not found
}

// read user credentials from a text file (JSON or YAML)
// no duplicates are allowed
func readUsersFile(fileName string) ([]*UserInfo, error) {
	// read whole file
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	users := make([]*UserInfo, 0)
	ext := filepath.Ext(fileName)
	switch strings.ToLower(ext) {
	case ".json":
		err = json.Unmarshal(data, &users)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &users)
	default:
		err = fmt.Errorf("%q is unknown credentials file extention", ext)
	}

	if err != nil {
		return nil, err
	}

	return users, nil // OK
}

// check for duplicated and build users map
func checkForDuplicates(users []*UserInfo) (map[string]*UserInfo, error) {
	// check for duplicates
	unique := make(map[string]*UserInfo, len(users))
	for _, u := range users {
		if unique[u.Name] != nil {
			return nil, fmt.Errorf("%q duplicate user info", u.Name)
		}
		unique[u.Name] = u
	}

	return unique, nil // OK
}

// ParseSecret string
func ParseSecret(secret string) ([]byte, error) {
	if strings.HasPrefix(secret, "@") {
		return ioutil.ReadFile(strings.TrimPrefix(secret, "@"))
	} else if strings.HasPrefix(secret, "base64:") {
		return base64.StdEncoding.DecodeString(strings.TrimPrefix(secret, "base64:"))
	} else if strings.HasPrefix(secret, "hex:") {
		return hex.DecodeString(strings.TrimPrefix(secret, "hex:"))
	}

	return []byte(secret), nil // OK
}
