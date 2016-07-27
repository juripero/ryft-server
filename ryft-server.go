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
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftdec"
	_ "github.com/getryft/ryft-server/search/ryfthttp"
	"github.com/getryft/ryft-server/search/ryftmux"
	_ "github.com/getryft/ryft-server/search/ryftone"
	_ "github.com/getryft/ryft-server/search/ryftprim"

	"github.com/getryft/ryft-server/middleware/auth"
	"github.com/getryft/ryft-server/middleware/cors"
	"github.com/getryft/ryft-server/middleware/gzip"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/thoas/stats"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
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

// Server instance
type Server struct {
	SearchBackend  string                 `yaml:"search-backend,omitempty"`
	BackendOptions map[string]interface{} `yaml:"backend-options,omitempty"`
	LocalOnly      bool                   `yaml:"local-only,omitempty"`
	DebugMode      bool                   `yaml:"debug-mode,omitempty"`
	KeepResults    bool                   `yaml:"keep-results,omitempty"`

	ListenAddress string `yaml:"address,omitempty"`
	listenAddress *net.TCPAddr

	TLS struct {
		Enabled       bool   `yaml:"enabled,omitempty"`
		ListenAddress string `yaml:"address,omitempty"`
		CertFile      string `yaml:"cert-file,omitempty"`
		KeyFile       string `yaml:"key-file,omitempty"`
	} `yaml:"tls,omitempty"`

	AuthType string `yaml:"auth-type,omitempty"`

	AuthFile struct {
		UsersFile string `yaml:"users-file,omitempty"`
	} `yaml:"auth-file,omitempty"`

	AuthLdap struct {
		Server   string `yaml:"server,omitempty"`
		Username string `yaml:"username,omitempty"`
		Password string `yaml:"password,omitempty"`
		Query    string `yaml:"query,omitempty"`
		BaseDN   string `yaml:"basedn,omitempty"`
	} `yaml:"auth-ldap,omitempty"`

	AuthJwt struct {
		Algorithm string `yaml:"algorithm,omitempty"`
		Secret    string `yaml:"secret,omitempty"`
		Lifetime  string `yaml:"lifetime,omitempty"`
	} `yaml:"auth-jwt,omitempty"`

	BusynessTolerance int `yaml:"busyness-tolerance,omitempty"`

	// the number of active search requests on this node
	// is used as a metric for "busyness"
	// worker thread is started if "local mode" is disabled
	activeSearchCount int32
	busynessChanged   chan int32
}

// config file name kingpin.Value
type serverConfigValue struct {
	s *Server
	v string
}

// set server's configuration file
func (f *serverConfigValue) Set(s string) error {
	f.v = s
	return f.s.parseConfig(f.v)
}

// get server's configuration file
func (f *serverConfigValue) String() string {
	return f.v
}

// create new server instance
func NewServer() (*Server, error) {
	s := new(Server)

	// default configuration
	s.SearchBackend = "ryftprim"
	s.BackendOptions = map[string]interface{}{}

	config := &serverConfigValue{s: s}
	kingpin.Flag("config", "Server configuration in YML format.").SetValue(config)
	kingpin.Flag("local-only", "Run server is local mode (no cluster).").BoolVar(&s.LocalOnly)
	kingpin.Flag("keep", "Keep temporary search result files.").Short('k').BoolVar(&s.KeepResults)
	kingpin.Flag("debug", "Run server in debug mode (more log messages).").Short('d').BoolVar(&s.DebugMode)
	kingpin.Flag("busyness-tolerance", "Cluster busyness tolerance.").Default("0").IntVar(&s.BusynessTolerance)

	kingpin.Flag("address", "Address:port to listen on.").Short('l').Default(":8765").StringVar(&s.ListenAddress)
	kingpin.Flag("tls", "Enable TLS/SSL.").Short('t').BoolVar(&s.TLS.Enabled)
	kingpin.Flag("tls-cert", "Certificate file. Required for --tls enabled.").StringVar(&s.TLS.CertFile)
	kingpin.Flag("tls-key", "Key-file. Required for --tls enabled.").StringVar(&s.TLS.KeyFile)
	kingpin.Flag("tls-address", "HTTPS address:port to listen on.").Default(":8766").StringVar(&s.TLS.ListenAddress)

	kingpin.Flag("auth", "Authentication type: none, file, ldap.").Short('a').Default("none").EnumVar(&s.AuthType, "none", "file", "ldap")
	kingpin.Flag("users-file", "User credentials filename. Required for --auth=file.").ExistingFileVar(&s.AuthFile.UsersFile)
	kingpin.Flag("jwt-alg", "JWT signing algorithm.").Default("HS256").StringVar(&s.AuthJwt.Algorithm)
	kingpin.Flag("jwt-secret", "JWT secret. Required for --auth=file or --auth=ldap.").StringVar(&s.AuthJwt.Secret)
	kingpin.Flag("jwt-lifetime", "JWT token lifetime.").Default("1h").StringVar(&s.AuthJwt.Lifetime)

	kingpin.Flag("ldap-server", "LDAP Server address:port. Required for --auth=ldap.").StringVar(&s.AuthLdap.Server)
	kingpin.Flag("ldap-user", "LDAP username for binding. Required for --auth=ldap.").StringVar(&s.AuthLdap.Username)
	kingpin.Flag("ldap-pass", "LDAP password for binding. Required for --auth=ldap.").StringVar(&s.AuthLdap.Password)
	kingpin.Flag("ldap-query", "LDAP user lookup query. Required for --auth=ldap.").Default("(&(uid=%s))").StringVar(&s.AuthLdap.Query)
	kingpin.Flag("ldap-basedn", "LDAP BaseDN for lookups. Required for --auth=ldap.").StringVar(&s.AuthLdap.BaseDN)

	kingpin.Parse()

	// check extra dependencies logic not handled by kingpin
	var err error

	switch strings.ToLower(s.AuthType) {
	case "file":
		switch {
		case len(s.AuthFile.UsersFile) == 0:
			kingpin.FatalUsage("users-file is required for file authentication.")
		case len(s.AuthJwt.Secret) == 0:
			kingpin.FatalUsage("jwt-secret is required for any authentication.")
		}

	case "ldap":
		switch {
		case len(s.AuthLdap.Server) == 0:
			kingpin.FatalUsage("ldap-server is required for ldap authentication.")
		case len(s.AuthLdap.Username) == 0:
			kingpin.FatalUsage("ldap-user is required for ldap authentication.")
		case len(s.AuthLdap.Password) == 0:
			kingpin.FatalUsage("ldap-pass is required for ldap authentication.")
		case len(s.AuthLdap.BaseDN) == 0:
			kingpin.FatalUsage("ldap-basedn is required for ldap authentication.")
		case len(s.AuthJwt.Secret) == 0:
			kingpin.FatalUsage("jwt-secret is required for any authentication.")
		}
	}

	if s.TLS.Enabled {
		switch {
		case len(s.TLS.ListenAddress) == 0:
			kingpin.FatalUsage("tls-address option is required for TLS enabled")
		case len(s.TLS.CertFile) == 0:
			kingpin.FatalUsage("tls-cert option is required for TLS enabled")
		case len(s.TLS.KeyFile) == 0:
			kingpin.FatalUsage("tls-key option is required for TLS enabled")
		}
	}

	if s.listenAddress, err = net.ResolveTCPAddr("tcp", s.ListenAddress); err != nil {
		kingpin.FatalUsage("%q is not a valid TCP address: %s", s.ListenAddress, err)
	}

	return s, nil // OK
}

// parse server configuration from YML file
func (s *Server) parseConfig(fileName string) error {
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
func (s *Server) getSearchEngine(localOnly bool, files []string, authToken, homeDir, userTag string) (search.Engine, error) {
	if !s.LocalOnly && !localOnly {
		return s.getClusterSearchEngine(files, authToken, homeDir, userTag)
	}

	return s.getLocalSearchEngine(homeDir)
}

// get cluster's search engine
func (s *Server) getClusterSearchEngine(files []string, authToken, homeDir, userTag string) (search.Engine, error) {
	// for each service create corresponding search engine
	services, tags, err := s.getConsulInfo(userTag, files)
	if err != nil {
		return nil, fmt.Errorf("failed to get consul services: %s", err)
	}
	log.WithField("tags", tags).WithField("services", services).Debug("cluster search")

	// if no tags required - use all nodes
	all_nodes := (len(tags) == 0)
	is_local := true // assume local service is used

	// list of tags required
	tags_required := make(map[string]bool)
	for _, t := range tags {
		tags_required[t] = true
	}

	// go through service tags and update `tags_required` map
	// return match count, matched tags are removed
	update_tags := func(serviceTags []string) int {
		count := 0
		for _, s := range serviceTags {
			if _, ok := tags_required[s]; ok {
				delete(tags_required, s)
				count += 1
			}
		}
		return count
	}

	// all services should be already arranged based on metrics
	backends := []search.Engine{}
	nodes := []string{}
	for _, service := range services {
		// stop if no more tags required
		if !all_nodes && len(tags_required) == 0 {
			break
		}

		// skip if no required tags found
		log.WithField("tags", service.ServiceTags).Debug("remote node tags")
		if !all_nodes && update_tags(service.ServiceTags) == 0 {
			continue // no tags found, skip this node
		}
		log.WithField("tags", tags_required).Debug("remain (remote) tags required")

		// remote node: use RyftHTTP backend
		port := service.ServicePort
		scheme := "http"
		var url string
		if port == 0 { // TODO: review the URL building!
			url = fmt.Sprintf("%s://%s:8765", scheme, service.Address)
		} else {
			url = fmt.Sprintf("%s://%s:%d", scheme, service.Address, port)
		}

		opts := map[string]interface{}{
			"server-url": url,
			"auth-token": authToken,
			"local-only": true,
			"skip-stat":  false,
			"index-host": url,
		}
		// log level
		if _, ok := opts["log-level"]; !ok && s.DebugMode {
			opts["log-level"] = "debug"
		}

		engine, err := search.NewEngine("ryfthttp", opts)
		if err != nil {
			return nil, err
		}
		backends = append(backends, engine)
		nodes = append(nodes, service.Node)
		if !s.isLocalService(service) {
			is_local = false
		}
	}

	// fail if there is remaining required tags
	if !all_nodes && len(tags_required) > 0 {
		rem := []string{} // remaining tags
		for k, _ := range tags_required {
			rem = append(rem, k)
		}
		return nil, fmt.Errorf("no services found for tags: %q", rem)
	}

	log.WithField("tags", tags).WithField("nodes", nodes).Infof("cluster search")

	if len(backends) > 0 && !is_local {
		engine, err := ryftmux.NewEngine(backends...)
		log.WithField("engine", engine).Debug("cluster search")
		return engine, err
	}

	// no services from consule, just use local search as a fallback
	log.Debugf("use local search as fallback")
	return s.getLocalSearchEngine(homeDir)
}

// get local search engine
func (s *Server) getLocalSearchEngine(homeDir string) (search.Engine, error) {
	opts := s.getBackendOptions()

	// some auto-options
	switch s.SearchBackend {
	case "ryftprim":
		// instance name
		if _, ok := opts["instance-name"]; !ok {
			opts["instance-name"] = fmt.Sprintf(".rest-%d", s.listenAddress.Port)
		}

		// home-dir (override settings)
		if _, ok := opts["home-dir"]; !ok || len(homeDir) > 0 {
			opts["home-dir"] = homeDir
		}

		// keep-files
		if _, ok := opts["keep-files"]; !ok {
			opts["keep-files"] = s.KeepResults
		}

		// index-host
		if _, ok := opts["index-host"]; !ok {
			opts["index-host"] = getHostName()
		}

		// log level
		if _, ok := opts["log-level"]; !ok && s.DebugMode {
			opts["log-level"] = "debug"
		}
	}

	backend, err := search.NewEngine(s.SearchBackend, opts)
	if err != nil {
		return backend, err
	}

	// special query decomposer
	if s.DebugMode {
		ryftdec.SetLogLevel("debug")
	}
	return ryftdec.NewEngine(backend)
}

// deep copy of backend options
func (s *Server) getBackendOptions() map[string]interface{} {
	opts := make(map[string]interface{})
	for k, v := range s.BackendOptions {
		opts[k] = v
	}
	return opts
}

// parse authentication token and home directory from context
func (s *Server) parseAuthAndHome(ctx *gin.Context) (userName string, authToken string, homeDir string, userTag string) {
	authToken = ctx.Request.Header.Get("Authorization") // may be empty

	// get home directory
	if v, exists := ctx.Get(gin.AuthUserKey); exists && v != nil {
		if user, ok := v.(*auth.UserInfo); ok {
			userName = user.Name
			homeDir = user.Home
			userTag = user.ClusterTag
		}
	}

	return
}

// update busyness thread
func (s *Server) startUpdatingBusyness() {
	s.busynessChanged = make(chan int32, 256)
	go func(metric int32) {
		var reported int32 = -1 // to force update metric ASAP
		for {
			select {
			case metric = <-s.busynessChanged:
				log.WithField("metric", metric).Debug("metric changed")
				continue
			case <-time.After(time.Second): // update latency
				if metric != reported {
					reported = metric
					log.WithField("metric", metric).Debug("metric reporting...")
					err := UpdateConsulMetric(int(metric))
					if err != nil {
						log.WithError(err).Warnf("failed to update consul metric")
					}
				}
				// TODO: graceful shutdown
			}
		}
	}(s.activeSearchCount)
}

// notify server a search is started
func (s *Server) onSearchStarted(config *search.Config) {
	metric := atomic.AddInt32(&s.activeSearchCount, +1)
	if s.busynessChanged != nil {
		// notify to update metric
		s.busynessChanged <- metric
	}
}

// notify server a search is started
func (s *Server) onSearchStopped(config *search.Config) {
	metric := atomic.AddInt32(&s.activeSearchCount, -1)
	if s.busynessChanged != nil {
		// notify to update metric
		s.busynessChanged <- metric
	}
}

// get local host name
func getHostName() string {
	hostName, _ := os.Hostname()
	return hostName
}

var serverStats = stats.New()

// RyftAPI include search, index, count
func main() {

	server, err := NewServer()
	if err != nil {
		log.WithError(err).Fatalf("Failed to read server configuration")
	}
	log.WithField("config", server).Infof("server configuration")

	// be quiet and efficient in production
	if !server.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	} else {
		log.Level = logrus.DebugLevel
	}

	// Create a rounter with default middleware: logger, recover
	router := gin.Default()

	// Logging & error recovery
	//	router.Use(gin.Logger())
	//	router.Use(srverr.Recovery())

	// Setting up Stats measurement middleware
	router.Use(func() gin.HandlerFunc {
		return func(c *gin.Context) {
			beginning := time.Now()
			c.Next()
			serverStats.End(beginning, stats.NewRecorderResponseWriter(c.Writer, http.StatusOK))
		}
	}())

	// Allow CORS requests for * (all domains)
	router.Use(cors.Cors("*"))

	// Enable GZip compression support
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// private endpoints
	private := router.Group("")

	// Enable authentication if configured
	var auth_provider auth.Provider
	switch strings.ToLower(server.AuthType) {
	case "file":
		file, err := auth.NewFile(server.AuthFile.UsersFile)
		if err != nil {
			log.WithError(err).Fatal("Failed to read users file")
		}
		auth_provider = file
	case "ldap":
		ldap, err := auth.NewLDAP(server.AuthLdap.Server, server.AuthLdap.Username,
			server.AuthLdap.Password, server.AuthLdap.Query, server.AuthLdap.BaseDN)
		if err != nil {
			log.WithError(err).Fatal("Failed to init LDAP authentication")
		}
		auth_provider = ldap
	case "none", "":
		break
	default:
		log.WithField("auth", server.AuthType).Fatalf("unknown authentication type")
	}

	// authentication enabled
	if auth_provider != nil {
		mw := auth.NewMiddleware(auth_provider, "")
		secret, err := auth.ParseSecret(server.AuthJwt.Secret)
		if err != nil {
			log.WithError(err).Fatalf("Failed to parse JWT secret")
		}
		lifetime, err := time.ParseDuration(server.AuthJwt.Lifetime)
		if err != nil {
			log.WithError(err).Fatalf("Failed to parse JWT lifetime")
		}
		mw.EnableJwt(secret, server.AuthJwt.Algorithm, lifetime)
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

	// stats page
	router.GET("/about", func(c *gin.Context) {
		c.JSON(http.StatusOK, serverStats.Data())
	})

	private.GET("/search", server.search)
	private.GET("/count", server.count)
	private.GET("/cluster/members", server.members)
	private.GET("/files", server.files)

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

	if !server.LocalOnly {
		server.startUpdatingBusyness()
	}

	// start listening on HTTP or HTTPS ports
	if server.TLS.Enabled {
		go router.RunTLS(server.TLS.ListenAddress, server.TLS.CertFile, server.TLS.KeyFile)
	}
	router.Run(server.ListenAddress)
}
