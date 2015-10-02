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
	LdapSettings struct {
		Port string
		Host  string
		Query string
		BindUsername string
		BindPassword string
	}
)

// BasicAuthForRealm returns a Basic HTTP Authorization middleware. It takes as arguments a map[string]string where
// the key is the user name and the value is the password, as well as the name of the Realm.
// If the realm is empty, "Authorization Required" will be used by default.
// (see http://tools.ietf.org/html/rfc2617#section-1.2)
func BasicAuthLDAPForRealm(settings LdapSettings, realm string) gin.HandlerFunc {

	if realm == "" {
		realm = "Authorization Required"
	}
	realm = "Basic realm=" + strconv.Quote(realm)

	return func(c *gin.Context) {

		if settings.Host == "" {
			setError(c, realm)
		} else if settings.Port == "" {
			setError(c, realm)
		}
		// Search user in the slice of allowed credentials
		user, binded := bindLDAP(settings, c.Request.Header.Get("Authorization"))

		if !binded {
			// Credentials doesn't match, we return 401 and abort handlers chain.
			fmt.Printf("\nAUTH: %v", "not binded\n")
			setError(c, realm)

		} else {
			fmt.Printf("\nAUTH: %v", "binded\n")
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
func BasicAuthLDAP(settings LdapSettings) gin.HandlerFunc {
	return BasicAuthLDAPForRealm(settings, "")
}

func authorizationHeader(user, password string) string {
	base := user + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(base))
}

func bindLDAP(settings LdapSettings, userdata string) (string, bool) {

	// The username and password we want to check
	username, password, ok := parseBasicAuth(userdata)

	if !ok {
		fmt.Printf("AUTH: couldn't parse '%v'\n", userdata)
		return "", false
	}

	// Connect to LDAP server
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", "ldap.forumsys.com", 389))
	if err != nil {
		log.Fatalf("Error Dialing LDAP: %v", err)
		return "", false
	}
	defer l.Close()

	// Reconnect with TLS
	err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		log.Fatalf("Error using TLS: %v", err)
		return "", false
	}

	log.Println("Binding..")
	// First bind with a read only user
	err = l.Bind(settings.BindUsername, settings.BindPassword)
	if err != nil {
		log.Fatalf("Error Binding readonly: %v", err)
		return "", false
	}

	if (settings.Query == ""){
		settings.Query = "(&(uid=%s))"
	}

	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		"dc=example,dc=com",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(settings.Query, username),
		[]string{"dn"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatalf("Error Searching: %v", err)
		return "", false
	}

	if len(sr.Entries) != 1 {
		log.Fatalf("User does not exist or too many entries returned: %v", searchRequest)
		return "", false
	}

	userdn := sr.Entries[0].DN

	// Bind as the user to verify their password
	err = l.Bind(userdn, password)
	if err != nil {
		log.Fatalf("Error binding User: %v", err)
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
