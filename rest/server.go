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

package rest

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftdec"
	_ "github.com/getryft/ryft-server/search/ryfthttp"
	"github.com/getryft/ryft-server/search/ryftmux"
	_ "github.com/getryft/ryft-server/search/ryftone"
	_ "github.com/getryft/ryft-server/search/ryftprim"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/catalog"

	"github.com/getryft/ryft-server/middleware/auth"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

var (
	// logger instance
	log = logrus.New()
)

// Server instance
type Server struct {
	SearchBackend  string                 `yaml:"search-backend,omitempty"`
	BackendOptions map[string]interface{} `yaml:"backend-options,omitempty"`
	LocalOnly      bool                   `yaml:"local-only,omitempty"`
	DebugMode      bool                   `yaml:"debug-mode,omitempty"`
	KeepResults    bool                   `yaml:"keep-results,omitempty"`

	ListenAddress       string `yaml:"address,omitempty"`
	ListenAddressParsed *net.TCPAddr

	HttpTimeout string `yaml:"http-timeout,omitempty"`

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

	AuthLdap auth.LdapConfig `yaml:"auth-ldap,omitempty"`

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

	BooleansPerExpression map[string]int `yaml:"booleans-per-expression"`

	// catalogs related options
	Catalogs struct {
		MaxDataFileSize  string `yaml:"max-data-file-size"`
		CacheDropTimeout string `yaml:"cache-drop-timeout"`
		DataDelimiter    string `yaml:"default-data-delim"`
		TempDirectory    string `yaml:"temp-dir"`
	} `yaml:"catalogs,omitempty"`

	// consul client is cached here
	consulClient interface{}
}

// create new server instance
func NewServer() *Server {
	s := new(Server)

	// default configuration
	s.SearchBackend = "ryftprim"
	s.BackendOptions = map[string]interface{}{}

	return s // OK
}

// set logging level
func SetLogLevel(level logrus.Level) {
	log.Level = level
}

// parse server configuration from YML file
func (s *Server) ParseConfig(fileName string) error {
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

	// validate catalog's maximum data file size
	if len(s.Catalogs.MaxDataFileSize) > 0 {
		if lim, err := utils.ParseDataSize(s.Catalogs.MaxDataFileSize); err != nil {
			return fmt.Errorf("failed to parse catalog maximum data size: %s", err)
		} else if lim == 0 {
			return fmt.Errorf("catalog maximum data size cannot be zero")
		} else {
			catalog.DefaultDataSizeLimit = lim
		}
	}

	// validate catalog's cache drop timeout
	if len(s.Catalogs.CacheDropTimeout) > 0 {
		if t, err := time.ParseDuration(s.Catalogs.CacheDropTimeout); err != nil {
			return fmt.Errorf("failed to parse catalog cache drop timeout: %s", err)
		} else {
			catalog.DefaultCacheDropTimeout = t
		}
	}

	// assign other catalog options
	catalog.DefaultDataDelimiter = s.Catalogs.DataDelimiter
	catalog.DefaultTempDirectory = s.Catalogs.TempDirectory

	return nil // OK
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
		log.WithField("service", service.Node).WithField("tags", service.ServiceTags).Debug("remote node tags")
		if !all_nodes && update_tags(service.ServiceTags) == 0 {
			continue // no tags found, skip this node
		}
		log.WithField("tags", tags_required).Debug("remain (remote) tags required")

		// use native search engine for local services!
		// (no sense to do extra HTTP call)
		if s.isLocalService(service) {
			engine, err := s.getLocalSearchEngine(homeDir)
			if err != nil {
				return nil, err
			}
			backends = append(backends, engine)
			nodes = append(nodes, service.Node)

			continue
		}

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
		is_local = false
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
			opts["instance-name"] = fmt.Sprintf(".rest-%d", s.ListenAddressParsed.Port)
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
	return ryftdec.NewEngine(backend, s.BooleansPerExpression, s.KeepResults)
}

// deep copy of backend options
func (s *Server) getBackendOptions() map[string]interface{} {
	opts := make(map[string]interface{})
	for k, v := range s.BackendOptions {
		opts[k] = v
	}
	return opts
}

// get read/write http timeout
func (s *Server) GetHttpTimeout() time.Duration {
	if len(s.HttpTimeout) == 0 {
		return 1 * time.Hour // default
	}

	d, err := time.ParseDuration(s.HttpTimeout)
	if err != nil {
		panic(err)
	}

	return d
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
func (s *Server) StartUpdatingBusyness() {
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
					err := s.updateConsulMetric(int(metric))
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
	s.onSearchChanged(config, +1)
}

// notify server a search is started
func (s *Server) onSearchStopped(config *search.Config) {
	s.onSearchChanged(config, -1)
}

// notify server a search is changed
func (s *Server) onSearchChanged(config *search.Config, delta int32) {
	metric := atomic.AddInt32(&s.activeSearchCount, delta)
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
