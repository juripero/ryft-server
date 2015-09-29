package auth

import (
	"encoding/base64"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	url = "ldap://ldap.forumsys.com:389/"
)

func AuthBasicLDAP() (gin.HandlerFunc, error) {

	// users, err := checkLDAP()
	// if err != nil {
	// 	return nil, err
	// }

	// return gin.BasicAuth(users), nil
	return nil, nil
}

// func checkLDAP() (map[string]string, error) {
//
// 	var user, passwd string
// 	// user = "cn=read-only-admin,dc=example,dc=com"
// 	// passwd = "password"
// 	userB64 := c.Request.Header.Get("Authorization")
//
// 	if userB64 == "" {
// 		// Credentials doesn't match, we return 401 and abort handlers chain.
// 		return nil, fmt.Errorf("No credantials were found")
// 	}
// 	user, passwd, ok := parseBasicAuth(userB64)
// 	if !ok {
// 		return nil, fmt.Errorf("Invalid credantials ")
// 	}
// 	// The user credentials was found, set user's id to key AuthUserKey in this context, the userId can be read later using
// 	c.Set(gin.AuthUserKey, user)
//
// 	ldap, err := openldap.Initialize(url)
// 	defer ldap.Close()
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	ldap.SetOption(openldap.LDAP_OPT_PROTOCOL_VERSION, openldap.LDAP_VERSION3)
// 	err = ldap.Bind(user, passwd)
//
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return gin.Accounts{user: passwd}, nil
// }

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
