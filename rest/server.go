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
	"path/filepath"
	"strings"
	"time"

	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/catalog"

	"github.com/getryft/ryft-server/middleware/auth"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

// ServerConfig server's configuration.
type ServerConfig struct {
	SearchBackend  string                 `yaml:"search-backend,omitempty"`
	BackendOptions map[string]interface{} `yaml:"backend-options,omitempty"`
	LocalOnly      bool                   `yaml:"local-only,omitempty"`
	DebugMode      bool                   `yaml:"debug-mode,omitempty"`
	KeepResults    bool                   `yaml:"keep-results,omitempty"`

	Logging        string                       `yaml:"logging,omitempty"`
	LoggingOptions map[string]map[string]string `yaml:"logging-options,omitempty"`

	ListenAddress string `yaml:"address,omitempty"`

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

	BooleansPerExpression map[string]int `yaml:"booleans-per-expression"`

	// catalogs related options
	Catalogs struct {
		MaxDataFileSize  string `yaml:"max-data-file-size"`
		CacheDropTimeout string `yaml:"cache-drop-timeout"`
		DataDelimiter    string `yaml:"default-data-delim"`
		TempDirectory    string `yaml:"temp-dir"`
	} `yaml:"catalogs,omitempty"`

	SettingsPath string `yaml:"settings-path,omitempty"`
}

// Server instance
type Server struct {
	Config ServerConfig

	listenAddress *net.TCPAddr

	// the number of active search requests on this node
	// is used as a metric for "busyness"
	// worker thread is started if "local mode" is disabled
	activeSearchCount int32
	busynessChanged   chan int32

	// consul client is cached here
	consulClient interface{}

	settings       *ServerSettings
	gotPendingJobs chan int // signal new jobs added
}

// create new server instance
func NewServer() *Server {
	s := new(Server)

	// default configuration
	s.Config.SearchBackend = "ryftprim"
	s.Config.BackendOptions = map[string]interface{}{}
	s.Config.SettingsPath = "/var/ryft/server.settings"

	return s // OK
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

	err = yaml.Unmarshal(buf, &s.Config)
	if err != nil {
		return fmt.Errorf("failed to parse configuration from %q: %s", fileName, err)
	}

	// validate catalog's maximum data file size
	if len(s.Config.Catalogs.MaxDataFileSize) > 0 {
		if lim, err := utils.ParseDataSize(s.Config.Catalogs.MaxDataFileSize); err != nil {
			return fmt.Errorf("failed to parse catalog maximum data size: %s", err)
		} else if lim == 0 {
			return fmt.Errorf("catalog maximum data size cannot be zero")
		} else {
			catalog.DefaultDataSizeLimit = lim
		}
	}

	// validate catalog's cache drop timeout
	if len(s.Config.Catalogs.CacheDropTimeout) > 0 {
		if t, err := time.ParseDuration(s.Config.Catalogs.CacheDropTimeout); err != nil {
			return fmt.Errorf("failed to parse catalog cache drop timeout: %s", err)
		} else {
			catalog.DefaultCacheDropTimeout = t
		}
	}

	// assign other catalog options
	catalog.DefaultDataDelimiter = s.Config.Catalogs.DataDelimiter
	catalog.DefaultTempDirectory = s.Config.Catalogs.TempDirectory

	return nil // OK
}

// apply configuration
func (s *Server) Prepare() (err error) {
	if s.listenAddress, err = net.ResolveTCPAddr("tcp", s.Config.ListenAddress); err != nil {
		return fmt.Errorf("%q is not a valid TCP address: %s", s.Config.ListenAddress, err)
	}

	// settings
	settingsDir, _ := filepath.Split(s.Config.SettingsPath)
	_ = os.MkdirAll(settingsDir, 0755)
	if s.settings, err = OpenSettings(s.Config.SettingsPath); err != nil {
		return fmt.Errorf("failed to open settings: %s", err)
	}

	// pending jobs
	s.gotPendingJobs = make(chan int, 256)
	go s.processPendingJobs()

	// automatic debug mode
	if len(s.Config.Logging) == 0 && s.Config.DebugMode {
		s.Config.Logging = "debug"

		// if no "debug" section, create it...
		if _, ok := s.Config.LoggingOptions[s.Config.Logging]; !ok {
			if s.Config.LoggingOptions == nil {
				s.Config.LoggingOptions = make(map[string]map[string]string)
			}
			s.Config.LoggingOptions[s.Config.Logging] = makeDefaultLoggingOptions("debug")
		}
	}

	// logging levels
	if len(s.Config.Logging) > 0 {
		if cfg, ok := s.Config.LoggingOptions[s.Config.Logging]; ok {
			for key, val := range cfg {
				if err := setLoggingLevel(key, val); err != nil {
					return fmt.Errorf("failed to apply logging level for '%s': %s", key, err)
				}
			}
		} else {
			return fmt.Errorf("no valid logging options found for '%s'", s.Config.Logging)
		}
	}

	// business update
	if !s.Config.LocalOnly {
		s.startUpdatingBusyness()
	}

	return nil // OK
}

// get read/write http timeout
func (s *Server) GetHttpTimeout() time.Duration {
	if len(s.Config.HttpTimeout) == 0 {
		return 1 * time.Hour // default
	}

	d, err := time.ParseDuration(s.Config.HttpTimeout)
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

// add new pending job
func (s *Server) addPendingJob(command, arguments string, when time.Time) {
	pjobLog.WithFields(map[string]interface{}{
		"command":   command,
		"arguments": arguments,
		"when":      when,
	}).Debugf("adding new pending job")

	s.settings.AddJob(command, arguments, when)
	s.gotPendingJobs <- 1 // notify processing goroutine about new job
	// TODO: do not notify many times
}

// process pending jobs
func (s *Server) processPendingJobs() {
	// sleep a while before start
	time.Sleep(1 * time.Second)

	for {
		now := time.Now()

		// get Job list to be done (1 second in advance)
		pjobLog.WithField("time", now).Debug("get pending jobs")
		jobs, err := s.settings.QueryAllJobs(now.Add(1 * time.Second))
		if err != nil {
			pjobLog.WithError(err).Warn("failed to get pending jobs")
			time.Sleep(10 * time.Second)
		}

		// do jobs
		ids := []int64{} // completed
		for job := range jobs {
			if s.doPendingJob(job) {
				ids = append(ids, job.Id)
			}
		}

		// delete completed jobs
		if len(ids) > 0 {
			pjobLog.WithField("jobs", ids).Debug("mark jobs as completed")
			if err = s.settings.DelJobs(ids); err != nil {
				log.WithError(err).Warn("failed to delete completed jobs")
			}
		}

		next, err := s.settings.GetNextJobTime()
		if err != nil {
			pjobLog.WithError(err).Warn("failed to get next job time")
			next = now.Add(1 * time.Hour)
		}
		pjobLog.WithField("time", next).Debug("next job time")

		sleep := next.Sub(now)
		if sleep < time.Second {
			sleep = time.Second
		}

		pjobLog.WithField("sleep", sleep).Debug("sleep a while before next step")
		select {
		case <-time.After(sleep):
			continue
		case <-s.gotPendingJobs:
			continue
		}
	}
}

// do pending job
func (s *Server) doPendingJob(job SettingsJobItem) bool {
	switch strings.ToLower(job.Cmd) {
	case "delete-file":
		res := deleteAll("/", []string{job.Args})
		pjobLog.WithFields(map[string]interface{}{
			"file":   job.Args,
			"result": res,
		}).Debug("pending job: delete file")
		return true

	case "delete-catalog":
		res := deleteAllCatalogs("/", []string{job.Args})
		pjobLog.WithFields(map[string]interface{}{
			"catalog": job.Args,
			"result":  res,
		}).Debug("pending job: delete catalog")
		return true
	}

	pjobLog.WithFields(map[string]interface{}{
		"command":   job.Cmd,
		"arguments": job.Args,
	}).Warn("unknown command, ignored")
	// return false // will be processed later
	return true // ignore job
}

// get local host name
func getHostName() string {
	hostName, _ := os.Hostname()
	return hostName
}
