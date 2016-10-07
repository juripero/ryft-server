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
	"time"

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

// ServerConfig server's configuration.
type ServerConfig struct {
	SearchBackend  string                 `yaml:"search-backend,omitempty"`
	BackendOptions map[string]interface{} `yaml:"backend-options,omitempty"`
	LocalOnly      bool                   `yaml:"local-only,omitempty"`
	DebugMode      bool                   `yaml:"debug-mode,omitempty"`
	KeepResults    bool                   `yaml:"keep-results,omitempty"`

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
}

// create new server instance
func NewServer() *Server {
	s := new(Server)

	// default configuration
	s.Config.SearchBackend = "ryftprim"
	s.Config.BackendOptions = map[string]interface{}{}

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
func (s *Server) ApplyConfig() (err error) {
	if s.listenAddress, err = net.ResolveTCPAddr("tcp", s.Config.ListenAddress); err != nil {
		return fmt.Errorf("%q is not a valid TCP address: %s", s.Config.ListenAddress, err)
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

// get local host name
func getHostName() string {
	hostName, _ := os.Hostname()
	return hostName
}
