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
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/getryft/ryft-server/search"
	_ "github.com/getryft/ryft-server/search/ryfthttp"
	"github.com/getryft/ryft-server/search/ryftmux"
	_ "github.com/getryft/ryft-server/search/ryftprim"

	"github.com/getryft/ryft-server/encoder"
	"github.com/getryft/ryft-server/middleware/auth"
	"github.com/getryft/ryft-server/middleware/cors"
	"github.com/getryft/ryft-server/middleware/gzip"
	"github.com/getryft/ryft-server/srverr"

	"github.com/gin-gonic/gin"
	"github.com/thoas/stats"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

var (
	serverConfig = kingpin.Flag("config", "Server configuration in YML format.").String()

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

// Server instance
type Server struct {
	SearchBackend  string                 `yaml:"searchBackend,omitempty"`
	BackendOptions map[string]interface{} `yaml:"backendOptions,omitempty"`
}

// parse server configuration from YML file
func (s *Server) parseConfig(fileName string) error {
	// default configuration if no file provided
	s.SearchBackend = "ryftprim"
	s.BackendOptions = map[string]interface{}{}

	if len(fileName) == 0 {
		return nil // OK
	}

	// read full file content
	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read configuration from %q: %s", fileName, err)
	}

	// TODO: parse ServerConfig dedicated structure
	err = yaml.Unmarshal(buf, &s)
	if err != nil {
		return fmt.Errorf("failed to parse configuration from %q: %s", fileName, err)
	}

	return nil // OK
}

func ensureDefault(flag *string, message string) {
	if *flag == "" {
		kingpin.FatalUsage(message)
	}
}

// get search backend with options
func (s *Server) getSearchEngine(localOnly bool) (search.Engine, error) {
	if !localOnly {
		// cluster search

		// for each service create corresponding search engine
		backends := []search.Engine{}
		info, err := GetConsulInfo()
		if err != nil {
			return nil, fmt.Errorf("failed to get consul service info: %s", err)
		}
		for _, service := range info {
			if compareIP(service.Address) && service.ServicePort == (*listenAddress).Port {
				// local node: just use normal backend
				engine, err := s.getSearchEngine(true)
				if err != nil {
					return nil, err
				}
				backends = append(backends, engine)
				continue // skip
			}

			// remote node: use RyftHTTP backend
			port := service.ServicePort
			scheme := "http"
			var url string
			if port == 0 { // TODO: review the URL building!
				url = fmt.Sprintf("%s://%s:%s", scheme, service.Address, DefaultPort)
			} else {
				url = fmt.Sprintf("%s://%s:%d", scheme, service.Address, port)
			}

			opts := map[string]interface{}{
				"server-url": url,
				"local-only": true,
				"skip-stat":  false,
				"index-host": url,
			}
			// log level
			if _, ok := opts["log-level"]; !ok && *debug {
				opts["log-level"] = "debug"
			}

			engine, err := search.NewEngine("ryfthttp", opts)
			if err != nil {
				return nil, err
			}
			backends = append(backends, engine)
		}

		if len(backends) > 0 {
			return ryftmux.NewEngine(backends...)
		} else {
			// no services from consule, just use local search as a fallback
			return s.getSearchEngine(true)
		}
	} else {
		// local node search
		opts := s.BackendOptions

		// some auto-options
		switch s.SearchBackend {
		case "ryftprim":
			// instance name
			if _, ok := opts["instance-name"]; !ok {
				opts["instance-name"] = fmt.Sprintf("RyftServer-%d", (*listenAddress).Port)
			}

			// keep-files
			if _, ok := opts["keep-files"]; !ok {
				opts["keep-files"] = *KeepResults
			}

			// index-host
			if _, ok := opts["index-host"]; !ok {
				hostName, _ := os.Hostname()
				opts["index-host"] = hostName
			}

			// log level
			if _, ok := opts["log-level"]; !ok && *debug {
				opts["log-level"] = "debug"
			}
		}

		return search.NewEngine(s.SearchBackend, opts)
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

var Stats = stats.New()

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

	// Create a rounter with default middleware: logger, recover
	router := gin.Default()

	// Configure required middlewares
	var server Server
	err := server.parseConfig(*serverConfig)
	if err != nil {
		log.Fatalf("Failed to read server configuration: %s", err)
	}
	log.Printf("CONFIG: %+v", server)

	// Logging & error recovery
	//	router.Use(gin.Logger())
	//	router.Use(srverr.Recovery())

	// Setting up Stats measirment middleware
	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			beginning := time.Now()
			c.Next()
			Stats.End(beginning, c.Writer)
		}
	}())

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

	// stats page
	router.GET("/about", func(c *gin.Context) {
		c.JSON(http.StatusOK, Stats.Data())
	})

	router.GET("/search", detectEncoder, server.search)
	router.GET("/count", detectEncoder, server.count)
	router.GET("/cluster/members", server.members)
	router.GET("/files", server.files)

	// static asset
	for _, asset := range AssetNames() {
		data := MustAsset(asset)
		ct := mime.TypeByExtension(filepath.Ext(asset))
		router.GET("/"+asset, func(c *gin.Context) {
			c.Data(http.StatusOK, ct, data)
		})
	}

	// index
	idxHTML := MustAsset("index.html")
	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, http.DetectContentType(idxHTML), idxHTML)
	})

	// Startup preparatory

	// start listening on HTTP or HTTPS ports
	if *tlsEnabled {
		go router.RunTLS((*tlsListenAddress).String(), *tlsCrtFile, *tlsKeyFile)
	}
	router.Run((*listenAddress).String())
}

const (
	ENCODER_CONTEXT_KEY = "encoder-detected"
)

func detectEncoder(c *gin.Context) {
	accept := c.NegotiateFormat(encoder.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = encoder.MIME_JSON
	}
	c.Header("Content-Type", accept)

	// setting up encoder to respond with requested format
	if enc, err := encoder.GetByMimeType(accept); err != nil {
		panic(srverr.New(http.StatusBadRequest, err.Error()))
	} else {
		c.Set(ENCODER_CONTEXT_KEY, enc)
	}
}

func encoderFromContext(c *gin.Context) encoder.Encoder {
	// TODO add handlers for null value and report 400 error
	return c.MustGet(ENCODER_CONTEXT_KEY).(encoder.Encoder)
}
