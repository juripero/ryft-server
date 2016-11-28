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

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftdec"
	_ "github.com/getryft/ryft-server/search/ryfthttp"
	"github.com/getryft/ryft-server/search/ryftmux"
	_ "github.com/getryft/ryft-server/search/ryftprim"
	"github.com/getryft/ryft-server/search/utils"
)

// get search backend with options
func (s *Server) getSearchEngine(localOnly bool, files []string, authToken, homeDir, userTag string) (search.Engine, error) {
	if !s.Config.LocalOnly && !localOnly {
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
	switch s.Config.SearchBackend {
	case "ryftprim", "ryftone":
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
			opts["keep-files"] = s.Config.KeepResults
		}

		// index-host
		if _, ok := opts["index-host"]; !ok {
			opts["index-host"] = s.Config.HostName
		}
	}

	backend, err := search.NewEngine(s.Config.SearchBackend, opts)
	if err != nil {
		return backend, err
	}

	return ryftdec.NewEngine(backend /*s.Config.BooleansPerExpression*/, -1, s.Config.KeepResults)
}

// deep copy of backend options
func (s *Server) getBackendOptions() map[string]interface{} {
	opts := make(map[string]interface{})
	for k, v := range s.Config.BackendOptions {
		opts[k] = v
	}
	return opts
}

// get mount point path from local search engine
func (s *Server) getMountPoint(homeDir string) (string, error) {
	engine, err := s.getLocalSearchEngine(homeDir)
	if err != nil {
		return "", err
	}

	opts := engine.Options()
	return utils.AsString(opts["ryftone-mount"])
}

// cancels results if not done
func cancelIfNotDone(res *search.Result) {
	if !res.IsDone() { // cancel processing
		if errors, records := res.Cancel(); errors > 0 || records > 0 {
			log.WithFields(map[string]interface{}{
				"errors":  errors,
				"records": records,
			}).Debugf("[%s]: some errors/records are ignored (panic recover)", CORE)
		}
	}
}
