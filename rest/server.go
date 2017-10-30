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
	"time"

	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/catalog"

	"github.com/getryft/ryft-server/middleware/auth"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

// TimeDuration is a wrapper on time.Duration to support YAML marshaling
type TimeDuration struct {
	val *time.Duration
}

// Bind binds the wrapper and any value
func NewTimeDuration(val *time.Duration) TimeDuration {
	return TimeDuration{val: val}
}

// UnmarshalYAML unmarshals time duration from string
func (td *TimeDuration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// get as string first
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	// parse
	if d, err := time.ParseDuration(s); err != nil {
		return err
	} else {
		*td.val = d
	}

	return nil // OK
}

// MarshalYAML marshals time duration to string
func (td *TimeDuration) MarshalYAML() (interface{}, error) {
	return td.val.String(), nil
}

// ServerConfig server's configuration.
type ServerConfig struct {
	SearchBackend  string                 `yaml:"search-backend,omitempty"`
	BackendOptions map[string]interface{} `yaml:"backend-options,omitempty"`
	LocalOnly      bool                   `yaml:"local-only,omitempty"`
	DebugMode      bool                   `yaml:"debug-mode,omitempty"`
	KeepResults    bool                   `yaml:"keep-results,omitempty"`
	ExtraRequest   bool                   `yaml:"extra-request,omitempty"`

	Logging        string                       `yaml:"logging,omitempty"`
	LoggingOptions map[string]map[string]string `yaml:"logging-options,omitempty"`

	ListenAddress string `yaml:"address,omitempty"`

	HttpTimeout_ TimeDuration  `yaml:"http-timeout,omitempty"`
	HttpTimeout  time.Duration `yaml:"-"`

	ShutdownTimeout_ TimeDuration  `yaml:"shutdown-timeout,omitempty"`
	ShutdownTimeout  time.Duration `yaml:"-"`

	ProcessingThreads int `yaml:"processing-threads"`

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

	Busyness struct {
		Tolerance      int           `yaml:"tolerance,omitempty"`
		UpdateLatency_ TimeDuration  `yaml:"update-latency,omitempty"`
		UpdateLatency  time.Duration `yaml:"-"`
	} `yaml:"busyness,omitempty"`

	// catalogs related options
	Catalogs struct {
		MaxDataFileSize   string        `yaml:"max-data-file-size"`
		CacheDropTimeout_ TimeDuration  `yaml:"cache-drop-timeout"`
		CacheDropTimeout  time.Duration `yaml:"-"`
		DataDelimiter     string        `yaml:"default-data-delim"`
		TempDirectory     string        `yaml:"temp-dir"`
	} `yaml:"catalogs,omitempty"`

	SettingsPath string `yaml:"settings-path,omitempty"`
	HostName     string `yaml:"hostname,omitempty"`

	Sessions struct {
		Algorithm string `yaml:"signing-algorithm,omitempty"`
		Secret    string `yaml:"secret,omitempty"`
		secret    []byte `yaml:"-"`
		// Lifetime  string `yaml:"lifetime,omitempty"`
	} `yaml:"sessions,omitempty"`

	// post-processing scripts/actions
	PostProcScripts map[string]struct {
		ExecPath []string `yaml:"path"`
	} `yaml:"post-processing-scripts,omitempty"`

	// docker options
	Docker struct {
		RunCmd  []string            `yaml:"run"`
		ExecCmd []string            `yaml:"exec"`
		Images  map[string][]string `yaml:"images"`
	} `yaml:"docker,omitempty"`

	DefaultUserConfig map[string]interface{} `yaml:"default-user-config"`

	// expose debug info in REST API response
	DebugInternals bool `yaml:"debug-internals"`
}

// Server instance
type Server struct {
	Config ServerConfig

	// lsiten address (parsed)
	listenAddress *net.TCPAddr

	// auth manager
	AuthManager auth.Manager

	// the number of active search requests on this node
	// is used as a metric for "busyness"
	// worker thread is started if "local mode" is disabled
	activeSearchCount int32
	busynessChanged   chan int32

	// consul client is cached here
	consulClient interface{}

	settings    *ServerSettings
	gotJobsChan chan int // signal new jobs added
	// newJobsCount int32    // atomic

	closeCh chan struct{} // close all
}

// NewServer creates new server instance
func NewServer() *Server {
	s := new(Server)

	// default configuration
	// s.Config.ExtraRequest = true
	s.Config.SearchBackend = "ryftprim"
	s.Config.BackendOptions = map[string]interface{}{}
	s.Config.Busyness.UpdateLatency = 1 * time.Second
	s.Config.Busyness.UpdateLatency_ = NewTimeDuration(&s.Config.Busyness.UpdateLatency)
	s.Config.HttpTimeout = 1 * time.Hour
	s.Config.HttpTimeout_ = NewTimeDuration(&s.Config.HttpTimeout)
	s.Config.ShutdownTimeout = 10 * time.Minute
	s.Config.ShutdownTimeout_ = NewTimeDuration(&s.Config.ShutdownTimeout)
	s.Config.Catalogs.CacheDropTimeout = 10 * time.Second
	s.Config.Catalogs.CacheDropTimeout_ = NewTimeDuration(&s.Config.Catalogs.CacheDropTimeout)
	s.Config.SettingsPath = "/var/ryft/server.settings"
	s.Config.Sessions.Algorithm = "HS256"
	s.Config.Sessions.Secret = "session-secret-key"

	s.closeCh = make(chan struct{})
	return s // OK
}

// Close() closes the server
func (s *Server) Close() {
	close(s.closeCh)
}

// ParseConfig parses server configuration from YML file
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

	// assign catalog options
	catalog.SetDefaultCacheDropTimeout(s.Config.Catalogs.CacheDropTimeout)
	catalog.DefaultDataDelimiter = s.Config.Catalogs.DataDelimiter
	catalog.DefaultTempDirectory = s.Config.Catalogs.TempDirectory

	return nil // OK
}

// Prepare applies configuration
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

	// parse session secret
	s.Config.Sessions.secret, err = auth.ParseSecret(s.Config.Sessions.Secret)
	if err != nil {
		return fmt.Errorf("failed to parse session secret: %s", err)
	}

	// hostname
	if len(s.Config.HostName) == 0 {
		if h, err := os.Hostname(); err != nil {
			return fmt.Errorf("failed to get hostname: %s", err)
		} else {
			s.Config.HostName = h
		}
	}

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

	// busyness update
	if !s.Config.LocalOnly {
		s.startUpdatingBusyness()
	}

	// pending jobs
	s.startJobsProcessing()

	return nil // OK
}

// parse authentication token and home directory from context
func (s *Server) parseAuthAndHome(ctx *gin.Context) (userName string, authToken string, homeDir string, userTag string) {
	authToken = ctx.Request.Header.Get("Authorization") // may be empty

	// get home directory
	if v, exists := ctx.Get(gin.AuthUserKey); exists && v != nil {
		if user, ok := v.(*auth.UserInfo); ok {
			userName = user.Name
			homeDir = user.HomeDir
			userTag = user.ClusterTag
		}
	}

	return
}
