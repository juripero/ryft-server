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
	"math/rand"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/demon-xxi/wildmatch"
	"github.com/gin-gonic/gin"

	consul "github.com/hashicorp/consul/api"
)

// handle /cluster/members endpoint: information about cluster's nodes
func (s *Server) members(c *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(c)

	info, _, err := s.getConsulInfo("", nil) // no user tag, no files

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
// the services are arranged based on "busyness" metric!
func (s *Server) getConsulInfo(userTag string, files []string) (services []*consul.CatalogService, tags []string, err error) {
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

	// arrange services based on node metrics
	metrics, err := getNodeMetrics(client)
	if err != nil {
		return services, tags, fmt.Errorf("failed to get node metrics: %s", err)
	}
	services = s.rearrangeServices(services, metrics, s.BusynessTolerance)

	return services, tags, err
}

// re-arrange services from less busy to most
func (s *Server) rearrangeServices(services []*consul.CatalogService, metrics map[string]int, tolerance int) []*consul.CatalogService {
	if tolerance < 0 {
		tolerance = 0
	}

	// split services into groups based on metrics and tolerance
	groups := map[int][]*consul.CatalogService{}
	for _, service := range services {
		groupId := metrics[service.Node] / (tolerance + 1)
		groups[groupId] = append(groups[groupId], service)
	}

	// for the same group just use random shuffle
	services = make([]*consul.CatalogService, 0, len(services))
	for _, group := range groups {
		// local node goes first!
		local, remote := s.splitToLocalAndRemote(group)
		if local != nil {
			services = append(services, local)
		}

		// remote nodes are randomly shuffled!
		for _, k := range rand.Perm(len(remote)) {
			services = append(services, remote[k])
		}
	}

	return services
}

// UpdateConsulMetric updates the node metric in the cluster
func UpdateConsulMetric(metric int) error {
	config := consul.DefaultConfig()
	// TODO: get some data from server's configuration
	config.Datacenter = "dc1"
	client, err := consul.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to get consul client: %s", err)
	}

	name, err := client.Agent().NodeName()
	if err != nil {
		return fmt.Errorf("failed to get node name: %s", err)
	}

	pair := new(consul.KVPair)
	pair.Key = filepath.Join("busyness", name)
	pair.Value = []byte(fmt.Sprintf("%d", metric))
	_, err = client.KV().Put(pair, nil)
	if err != nil {
		return fmt.Errorf("failed to update node metric: %s", err)
	}

	return nil // OK
}

// get metric for all nodes
func getNodeMetrics(client *consul.Client) (map[string]int, error) {
	// get all wildcards (keys) and tags
	prefix := "busyness/"
	pairs, _, err := client.KV().List(prefix, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics from KV: %s", err)
	}

	metrics := map[string]int{}
	for _, kvp := range pairs {
		key, _ := url.QueryUnescape(kvp.Key)
		node := strings.TrimPrefix(key, prefix)
		metric, _ := strconv.ParseInt(string(kvp.Value), 10, 32)
		metrics[node] = int(metric)
	}

	return metrics, nil
}

// splits services to local and remote set
// NOTE the input `services` slice might be modified!
func (s *Server) splitToLocalAndRemote(services []*consul.CatalogService) (local *consul.CatalogService, remotes []*consul.CatalogService) {
	for i, service := range services {
		if s.isLocalService(service) {
			local = service
			remotes = append(services[:i],
				services[i+1:]...)
			return
		}
	}

	return nil, services // no local found
}

// check if service is local
func (s *Server) isLocalService(service *consul.CatalogService) bool {
	if compareIP(service.Address) && service.ServicePort == s.listenAddress.Port {
		return true
	}

	return false
}

// find best matched service tags for the file list
// userTag is used for multitenancy support
func findBestMatch(client *consul.Client, userTag string, files []string) ([]string, error) {
	if len(files) == 0 {
		return nil, nil // no files - no tags
	}

	// get all wildcards (keys) and tags
	prefix := filepath.Join(userTag, "partitions") + "/"
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
