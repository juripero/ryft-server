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
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/getryft/ryft-server/middleware/auth"
	"github.com/getryft/ryft-server/middleware/cors"
	"github.com/getryft/ryft-server/middleware/gzip"
	"github.com/getryft/ryft-server/rest"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	// logger instance
	log = logrus.New()
)

// customized via Makefile
var (
	Version string
	GitHash string
)

// config file name kingpin.Value
type serverConfigValue struct {
	s *rest.Server
	v string
}

// set server's configuration file
func (f *serverConfigValue) Set(s string) error {
	f.v = s
	return f.s.ParseConfig(f.v)
}

// get server's configuration file
func (f *serverConfigValue) String() string {
	return f.v
}

// RyftAPI include search, index, count
func main() {
	server := rest.NewServer()

	config := &serverConfigValue{s: server}
	kingpin.Flag("config", "Server configuration in YML format.").SetValue(config)
	kingpin.Flag("local-only", "Run server is local mode (no cluster).").BoolVar(&server.Config.LocalOnly)
	kingpin.Flag("keep", "Keep temporary search result files.").Short('k').BoolVar(&server.Config.KeepResults)
	kingpin.Flag("debug", "Run server in debug mode (more log messages).").Short('d').BoolVar(&server.Config.DebugMode)
	kingpin.Flag("logging", "Fine-tuned logging levels.").StringVar(&server.Config.Logging)
	kingpin.Flag("busyness-tolerance", "Cluster busyness tolerance.").Default("0").IntVar(&server.Config.BusynessTolerance)

	kingpin.Flag("address", "Address:port to listen on.").Short('l').Default(":8765").StringVar(&server.Config.ListenAddress)
	kingpin.Flag("tls", "Enable TLS/SSL.").Short('t').BoolVar(&server.Config.TLS.Enabled)
	kingpin.Flag("tls-cert", "Certificate file. Required for --tls enabled.").StringVar(&server.Config.TLS.CertFile)
	kingpin.Flag("tls-key", "Key-file. Required for --tls enabled.").StringVar(&server.Config.TLS.KeyFile)
	kingpin.Flag("tls-address", "HTTPS address:port to listen on.").Default(":8766").StringVar(&server.Config.TLS.ListenAddress)

	kingpin.Flag("auth", "Authentication type: none, file, ldap.").Short('a').Default("none").EnumVar(&server.Config.AuthType, "none", "file", "ldap")
	kingpin.Flag("users-file", "User credentials filename. Required for --auth=file.").ExistingFileVar(&server.Config.AuthFile.UsersFile)
	kingpin.Flag("jwt-alg", "JWT signing algorithm.").Default("HS256").StringVar(&server.Config.AuthJwt.Algorithm)
	kingpin.Flag("jwt-secret", "JWT secret. Required for --auth=file or --auth=ldap.").StringVar(&server.Config.AuthJwt.Secret)
	kingpin.Flag("jwt-lifetime", "JWT token lifetime.").Default("1h").StringVar(&server.Config.AuthJwt.Lifetime)

	kingpin.Flag("ldap-server", "LDAP Server address:port. Required for --auth=ldap.").StringVar(&server.Config.AuthLdap.ServerAddress)
	kingpin.Flag("ldap-user", "LDAP username for binding. Required for --auth=ldap.").StringVar(&server.Config.AuthLdap.BindUsername)
	kingpin.Flag("ldap-pass", "LDAP password for binding. Required for --auth=ldap.").StringVar(&server.Config.AuthLdap.BindPassword)
	kingpin.Flag("ldap-query", "LDAP user lookup query. Required for --auth=ldap.").Default("(&(uid=%s))").StringVar(&server.Config.AuthLdap.QueryFormat)
	kingpin.Flag("ldap-basedn", "LDAP BaseDN for lookups. Required for --auth=ldap.").StringVar(&server.Config.AuthLdap.BaseDN)

	kingpin.Parse()

	// check extra dependencies logic not handled by kingpin
	switch strings.ToLower(server.Config.AuthType) {
	case "file":
		switch {
		case len(server.Config.AuthFile.UsersFile) == 0:
			kingpin.FatalUsage("users-file is required for file authentication.")
		case len(server.Config.AuthJwt.Secret) == 0:
			kingpin.FatalUsage("jwt-secret is required for any authentication.")
		}

	case "ldap":
		switch {
		case len(server.Config.AuthLdap.ServerAddress) == 0:
			kingpin.FatalUsage("ldap-server is required for ldap authentication.")
		case len(server.Config.AuthLdap.BindUsername) == 0:
			kingpin.FatalUsage("ldap-user is required for ldap authentication.")
		case len(server.Config.AuthLdap.BindPassword) == 0:
			kingpin.FatalUsage("ldap-pass is required for ldap authentication.")
		case len(server.Config.AuthLdap.BaseDN) == 0:
			kingpin.FatalUsage("ldap-basedn is required for ldap authentication.")
		case len(server.Config.AuthJwt.Secret) == 0:
			kingpin.FatalUsage("jwt-secret is required for any authentication.")
		}
	}

	if server.Config.TLS.Enabled {
		switch {
		case len(server.Config.TLS.ListenAddress) == 0:
			kingpin.FatalUsage("tls-address option is required for TLS enabled")
		case len(server.Config.TLS.CertFile) == 0:
			kingpin.FatalUsage("tls-cert option is required for TLS enabled")
		case len(server.Config.TLS.KeyFile) == 0:
			kingpin.FatalUsage("tls-key option is required for TLS enabled")
		}
	}

	if err := server.Prepare(); err != nil {
		log.WithError(err).Fatal("failed to prepare server configuration")
	}

	log.WithField("version", Version).
		WithField("git-hash", GitHash).
		Infof("starting server...")
	log.WithField("config", server.Config).
		Debugf("server configuration")

	// be quiet and efficient in production
	if !server.Config.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	} else {
		rest.SetLogLevel(logrus.DebugLevel)
		log.Level = logrus.DebugLevel
	}

	// Create a rounter with default middleware: logger, recover
	router := gin.Default()

	// Allow CORS requests for * (all domains)
	router.Use(cors.Cors("*"))

	// Enable GZip compression
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// private endpoints
	private := router.Group("")

	// Enable authentication if configured
	var auth_provider auth.Provider
	switch strings.ToLower(server.Config.AuthType) {
	case "file":
		file, err := auth.NewFile(server.Config.AuthFile.UsersFile)
		if err != nil {
			log.WithError(err).Fatal("Failed to read users file")
		}
		auth_provider = file
	case "ldap":
		ldap, err := auth.NewLDAP(server.Config.AuthLdap)
		if err != nil {
			log.WithError(err).Fatal("Failed to init LDAP authentication")
		}
		auth_provider = ldap
	case "none", "":
		break
	default:
		log.WithField("auth", server.Config.AuthType).Fatal("unknown authentication type")
	}

	// authentication enabled
	if auth_provider != nil {
		mw := auth.NewMiddleware(auth_provider, "")
		secret, err := auth.ParseSecret(server.Config.AuthJwt.Secret)
		if err != nil {
			log.WithError(err).Fatal("Failed to parse JWT secret")
		}
		lifetime, err := time.ParseDuration(server.Config.AuthJwt.Lifetime)
		if err != nil {
			log.WithError(err).Fatal("Failed to parse JWT lifetime")
		}
		mw.EnableJwt(secret, server.Config.AuthJwt.Algorithm, lifetime)
		private.Use(mw.Authentication())
		private.GET("/token/refresh", mw.RefreshHandler())
		router.POST("/login", mw.LoginHandler())
	}

	router.GET("/version", func(ctx *gin.Context) {
		info := map[string]interface{}{
			"version":  Version,
			"git-hash": GitHash,
		}
		ctx.JSON(http.StatusOK, info)
	})

	private.GET("/search", server.DoSearch)
	private.GET("/count", server.DoCount)
	private.GET("/cluster/members", server.DoClusterMembers)
	private.GET("/files", server.DoGetFiles)
	private.DELETE("/files", server.DoDeleteFiles)
	private.POST("/files", server.DoPostFiles)

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

	if !server.Config.LocalOnly {
		server.StartUpdatingBusyness()
	}

	// start listening on HTTPS port
	if server.Config.TLS.Enabled {
		https_ep := &http.Server{Addr: server.Config.TLS.ListenAddress, Handler: router}
		https_ep.ReadTimeout = server.GetHttpTimeout()
		https_ep.WriteTimeout = server.GetHttpTimeout()

		go func() {
			if err := https_ep.ListenAndServeTLS(server.Config.TLS.CertFile, server.Config.TLS.KeyFile); err != nil {
				log.WithError(err).WithField("port", server.Config.TLS.ListenAddress).Fatal("failed to listen HTTPS")
			}
		}()
	}

	// start listening on HTTP port
	http_ep := &http.Server{Addr: server.Config.ListenAddress, Handler: router}
	http_ep.ReadTimeout = server.GetHttpTimeout()
	http_ep.WriteTimeout = server.GetHttpTimeout()
	if err := http_ep.ListenAndServe(); err != nil {
		log.WithError(err).WithField("port", server.Config.ListenAddress).Fatal("failed to listen HTTP")
	}
}
