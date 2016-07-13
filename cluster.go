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
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/demon-xxi/wildmatch"
	"github.com/gin-gonic/gin"

	consul "github.com/hashicorp/consul/api"
)

// handle /cluster/members endpoint: information about cluster's nodes
func (s *Server) members(c *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(c)

	info, _, err := GetConsulInfo("", nil) // no user tag, no files

	if err != nil {
		panic(NewServerError(http.StatusInternalServerError, err.Error()))
	} else {
		log.WithField("info", info).Debug("consul information")
		c.JSON(http.StatusOK, info)
	}
}

//type Service struct {
//	Node           string   `json:"Node"`
//	Address        string   `json:"Address"`
//	ServiceID      string   `json:"ServiceID"`
//	ServiceName    string   `json:"ServiceName"`
//	ServiceAddress string   `json:"ServiceAddress"`
//	ServiceTags    []string `json:"ServiceTags"`
//	ServicePort    string   `json:"ServicePort"`
//}

// GetConsulInfo gets the list of ryft services and
// the service tags related to requested set of files.
func GetConsulInfo(userTag string, files []string) (services []*consul.CatalogService, tags []string, err error) {
	config := consul.DefaultConfig()
	// TODO: get some data from server's configuration
	config.Datacenter = "dc1"
	client, err := consul.NewClient(config)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to get consul client: %s", err)
	}

	catalog := client.Catalog()
	services, _, err = catalog.Service("ryft-rest-api", "", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get consul services: %s", err)
	}

	if len(files) != 0 {
		tags, err = findBestMatch(client, userTag, files)
		if err != nil {
			return services, nil, fmt.Errorf("failed to get match tags: %s", err)
		}
	}

	return services, tags, err
}

// SplitToLocalAndRemote splits services to local and remote set
// NOTE the input `services` slice might be modified!
func SplitToLocalAndRemote(services []*consul.CatalogService) (local *consul.CatalogService, remotes []*consul.CatalogService) {
	for i, service := range services {
		if compareIP(service.Address) && service.ServicePort == (*listenAddress).Port {
			local = service
			remotes = append(services[:i],
				services[i+1:]...)
			return
		}
	}

	return nil, services // no local found
}

// find best matched service tags for the file list
// userTag is used for multitenancy support
func findBestMatch(client *consul.Client, userTag string, files []string) ([]string, error) {
	if len(files) == 0 {
		return nil, nil // no files - no tags
	}

	// get all wildcards (keys) and tags
	prefix := filepath.Join("partitions/", userTag)
	pairs, _, err := client.KV().List(prefix, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags from KV: %s", err)
	}

	keys := make([]string, len(pairs))
	tags := make([][]string, len(pairs))
	for i, kvp := range pairs {
		mask, _ := url.QueryUnescape(kvp.Key)
		keys[i] = strings.TrimPrefix(mask, prefix)
		tags[i] = strings.Split(string(kvp.Value), ",")

		// trim spaces from tags
		for k := range tags[i] {
			tags[i][k] = strings.TrimSpace(tags[i][k])
		}
	}

	// match files and wildcards
	tags_map := make(map[string]int)
	for _, f := range files {
		if found := wildmatch.IsSubsetOfAny(f, keys...); found >= 0 {
			for _, tag := range tags[found] {
				tags_map[tag] += 1
			}
		}
	}

	// map keys -> slice
	res := make([]string, 0, len(tags_map))
	for k := range tags_map {
		res = append(res, k)
	}

	return res, nil
}
