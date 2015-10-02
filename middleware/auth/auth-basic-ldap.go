package auth

////////
// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"

	"crypto/tls"
	"github.com/gin-gonic/gin"
	"gopkg.in/ldap.v2"
)

const AuthUserKey = "user"

type (
	ldapSettings struct {
		Address string
		Query string
		BindUsername string
		BindPassword string
		BaseDN string
	}
)

// BasicAuthForRealm returns a Basic HTTP Authorization middleware. It takes as arguments a map[string]string where
// the key is the user name and the value is the password, as well as the name of the Realm.
// If the realm is empty, "Authorization Required" will be used by default.
// (see http://tools.ietf.org/html/rfc2617#section-1.2)
func BasicAuthLDAPForRealm(settings ldapSettings, realm string) gin.HandlerFunc {

	if realm == "" {
		realm = "Authorization Required"
	}
	realm = "Basic realm=" + strconv.Quote(realm)

	return func(c *gin.Context) {

		// Search user in the slice of allowed credentials
		user, binded := bindLDAP(settings, c.Request.Header.Get("Authorization"))

		if !binded {
			// Credentials doesn't match, we return 401 and abort handlers chain.
//			log.Printf("\nAUTH: %v", "not binded\n")
			setError(c, realm)

		} else {
//			log.Printf("\nAUTH: %v", "binded\n")
			// The user credentials was found, set user's id to key AuthUserKey in this context, the userId can be read later using
			// c.MustGet(gin.AuthUserKey)
			c.Set(AuthUserKey, user)
		}
	}
}

func setError(c *gin.Context, realm string) {
	c.Header("WWW-Authenticate", realm)
	c.AbortWithStatus(401)
}

// BasicAuth returns a Basic HTTP Authorization middleware. It takes as argument a map[string]string where
// the key is the user name and the value is the password.
func BasicAuthLDAP(address, username, password, query, baseDN string) gin.HandlerFunc {
	settings := ldapSettings {
		address,
		query,
		username,
		password,
		baseDN,
	}
	return BasicAuthLDAPForRealm(settings, "")
}

func authorizationHeader(user, password string) string {
	base := user + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(base))
}

func bindLDAP(settings ldapSettings, userdata string) (string, bool) {

//	log.Printf("LDAP Settings: %+v", settings)

	// The username and password we want to check
	username, password, ok := parseBasicAuth(userdata)

	if !ok {
		log.Printf("AUTH: couldn't parse '%v'\n", userdata)
		return "", false
	}

	// Connect to LDAP server
	l, err := ldap.Dial("tcp", settings.Address)
	if err != nil {
		log.Printf("Error Dialing LDAP: %v", err)
		return "", false
	}
	defer l.Close()

	// Reconnect with TLS
	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		log.Printf("Error using TLS: %v", err)
		return "", false
	}

	// First bind with a read only user
	err = l.Bind(settings.BindUsername, settings.BindPassword)
	if err != nil {
		log.Printf("Error Binding readonly: %v", err)
		return "", false
	}

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		settings.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(settings.Query, username),
		[]string{"dn"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Printf("Error Searching: %v", err)
		return "", false
	}

	if len(sr.Entries) != 1 {
		log.Printf("User does not exist or too many entries returned: %v", searchRequest)
		return "", false
	}

	userdn := sr.Entries[0].DN

	// Bind as the user to verify their password
	err = l.Bind(userdn, password)
	if err != nil {
		log.Printf("Error binding User: %v", err)
		return "", false
	}

	return "authorizationHeader(username, password)", true
}

func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

////////
