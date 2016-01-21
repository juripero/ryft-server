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
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/getryft/ryft-server/middleware/auth"
	"github.com/getryft/ryft-server/middleware/cors"
	"github.com/getryft/ryft-server/middleware/gzip"
	"github.com/getryft/ryft-server/names"

	"github.com/gin-gonic/gin"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// KeepResults console flag for keeping results files
	KeepResults = kingpin.Flag("keep", "Keep search results temporary files.").Short('k').Bool()
	debug       = kingpin.Flag("debug", "Run http server in debug mode.").Short('d').Bool()

	authType      = kingpin.Flag("auth", "Authentication type: none, file, ldap.").Short('a').Enum("none", "file", "ldap")
	authUsersFile = kingpin.Flag("users-file", "File with user credentials. Required for --auth=file.").ExistingFile()

	authLdapServer = kingpin.Flag("ldap-server", "LDAP Server address:port. Required for --auth=ldap.").TCP()
	authLdapUser   = kingpin.Flag("ldap-user", "LDAP username for binding. Required for --auth=ldap.").String()
	authLdapPass   = kingpin.Flag("ldap-pass", "LDAP password for binding. Required for --auth=ldap.").String()
	authLdapQuery  = kingpin.Flag("ldap-query", "LDAP user lookup query. Defauls is '(&(uid=%s))'. Required for --auth=ldap.").Default("(&(uid=%s))").String()
	authLdapBase   = kingpin.Flag("ldap-basedn", "LDAP BaseDN for lookups.'. Required for --auth=ldap.").String()

	listenAddress = kingpin.Arg("address", "Address:port to listen on. Default is 0.0.0.0:8765.").Default("0.0.0.0:8765").TCP()

	tlsEnabled       = kingpin.Flag("tls", "Enable TLS/SSL. Default 'false'.").Short('t').Bool()
	tlsCrtFile       = kingpin.Flag("tls-crt", "Certificate file. Required for --tls=true.").ExistingFile()
	tlsKeyFile       = kingpin.Flag("tls-key", "Key-file. Required for --tls=true.").ExistingFile()
	tlsListenAddress = kingpin.Flag("tls-address", "Address:port to listen on HTTPS. Default is 0.0.0.0:8766").Default("0.0.0.0:8766").TCP()
)

func ensureDefault(flag *string, message string) {
	if *flag == "" {
		kingpin.FatalUsage(message)
	}
}

func parseParams() {
	kingpin.Parse()
	// check extra dependencies logic not handled by kingpin
	switch *authType {
	case "file":
		ensureDefault(authUsersFile, "users-file is required for file authentication.")
		break
	case "ldap":
		if (*authLdapServer) == nil {
			kingpin.FatalUsage("ldap-server is required for ldap authentication.")
		}
		if (*authLdapServer).IP == nil {
			kingpin.FatalUsage("ldap-server requires addresse name part, not only port.")
		}

		ensureDefault(authLdapUser, "ldap-user is required for ldap authentication.")
		ensureDefault(authLdapPass, "ldap-pass is required for ldap authentication.")
		ensureDefault(authLdapBase, "ldap-basedn is required for ldap authentication.")

		break
	}
	if *tlsEnabled {
		ensureDefault(tlsCrtFile, "tls-crt is required for enabled tls property")
		ensureDefault(tlsKeyFile, "tls-key is required for enabled tls property")
	}
}

// RyftAPI include search, index, count
func main() {

	// set log timestamp format
	log.SetFlags(log.Lmicroseconds)

	// parse command line arguments
	parseParams()

	// be quiet and efficient in production
	if !*debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Configure requred middlewares

	// Allow CORS requests for * (all domains)
	router.Use(cors.Cors("*"))

	// Enable GZip compression support
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// Enable authentication if configured
	switch *authType {
	case "file":
		auth, err := auth.AuthBasicFile(*authUsersFile)
		if err != nil {
			log.Printf("Error reading users file: %v", err)
			os.Exit(1)
		}
		router.Use(auth)
		break
	case "ldap":
		router.Use(auth.BasicAuthLDAP((*authLdapServer).String(), *authLdapUser,
			*authLdapPass, *authLdapQuery, *authLdapBase))
		break
	}

	// Configure routes

	// index & help
	idxHTML := MustAsset("index.html")
	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, http.DetectContentType(idxHTML), idxHTML)
	})

	// swagger schema
	swaggerJSON := MustAsset("swagger.json")
	router.GET("/swagger.json", func(c *gin.Context) {
		c.Data(http.StatusOK, http.DetectContentType(swaggerJSON), swaggerJSON)
	})

	// search method
	router.GET("/search", search)

	// count method
	router.GET("/count", count)

	// cluster members
	router.GET("/cluster/members", members)

	// Startup preparatory
	// Clean previously created folder
	names.Port = (*listenAddress).Port
	if err := os.RemoveAll(names.ResultsDirPath()); err != nil {
		log.Printf("Could not delete %s with error %s", names.ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	// Create folder for results cache
	if err := os.MkdirAll(names.ResultsDirPath(), 0777); err != nil {
		log.Printf("Could not create directory %s with error %s", names.ResultsDirPath(), err.Error())
		os.Exit(1)
	}

	// Name Generator will produce unique file names for each new results files
	names.StartNamesGenerator()

	// start listening on HTTP or HTTPS ports
	if *tlsEnabled {
		go router.RunTLS((*tlsListenAddress).String(), *tlsCrtFile, *tlsKeyFile)
	}
	router.Run((*listenAddress).String())
}
