package auth

////////
// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hamano/golang-openldap"
)

const AuthUserKey = "user"

type (
	Settings struct {
		Port string
		Url  string
		User UserDN
	}

	UserDN struct {
		UserPrefix  string
		UserPostfix string
	}
)

// BasicAuthForRealm returns a Basic HTTP Authorization middleware. It takes as arguments a map[string]string where
// the key is the user name and the value is the password, as well as the name of the Realm.
// If the realm is empty, "Authorization Required" will be used by default.
// (see http://tools.ietf.org/html/rfc2617#section-1.2)
func BasicAuthLDAPForRealm(settings Settings, realm string) gin.HandlerFunc {

	if realm == "" {
		realm = "Authorization Required"
	}
	realm = "Basic realm=" + strconv.Quote(realm)

	return func(c *gin.Context) {

		if settings.Url == "" {
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
func BasicAuthLDAP(settings Settings) gin.HandlerFunc {
	return BasicAuthLDAPForRealm(settings, "")
}

func authorizationHeader(user, password string) string {
	base := user + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(base))
}

func bindLDAP(settings Settings, userdata string) (string, bool) {
	username, password, ok := parseBasicAuth(userdata)

	if !ok {
		fmt.Printf("AUTH: %v", "couldn't parse\n")
		return "", false
	}

	url := settings.Url + ":" + settings.Port + "/"
	fmt.Printf("AUTH: %v", url+"\n")
	// user := settings.User.UserPrefix + username + "," + settings.User.UserPostfix
	user := "cn=" + username + "," + "dc=example,dc=com"

	ldap, err := openldap.Initialize(url)
	defer ldap.Close()

	if err != nil {
		return "", false
	}

	ldap.SetOption(openldap.LDAP_OPT_PROTOCOL_VERSION, openldap.LDAP_VERSION3)
	err = ldap.Bind(user, password)

	if err != nil {
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
