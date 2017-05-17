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
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	"github.com/demon-xxi/wildmatch"
	"github.com/gin-gonic/gin"

	consul "github.com/hashicorp/consul/api"
)

// handle /cluster/members endpoint: information about cluster's nodes
func (server *Server) DoClusterMembers(ctx *gin.Context) {
	// recover from panics if any
	defer RecoverFromPanic(ctx)

	services, _, err := server.getConsulInfo("", nil) // no user tag, no files
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()))
	}

	// convert services (only subset of fields)
	info := make([]map[string]interface{}, len(services))
	for i, s := range services {
		info[i] = map[string]interface{}{
			"node": s.Node,
			"tags": s.ServiceTags,
			"address": func() string {
				if s.ServicePort != 0 {
					return fmt.Sprintf("%s:%d", s.ServiceAddress, s.ServicePort)
				}
				return s.ServiceAddress
			}(),
		}
	}

	log.WithField("info", info).Info("cluster information")
	ctx.JSON(http.StatusOK, info)

}

// get consul client
func (server *Server) getConsulClient() (*consul.Client, error) {
	if server.consulClient != nil {
		return server.consulClient.(*consul.Client), nil // cached
	}

	// create new client
	config := consul.DefaultConfig()
	// TODO: get some data from server's configuration?
	config.Datacenter = "dc1"
	client, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}

	server.consulClient = client // put to cache
	return client, nil           // OK
}

// GetConsulInfo gets the list of ryft services
func (s *Server) getConsulInfoForFiles(userTag string, files []string) (services []*consul.CatalogService, tags [][]string, err error) {
	client, err := s.getConsulClient()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get consul client: %s", err)
	}

	catalog := client.Catalog()
	services, _, err = catalog.Service("ryft-rest-api", "", nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get consul services: %s", err)
	}

	tags, err = findAllMatches(client, userTag, files)
	if err != nil {
		return services, nil, fmt.Errorf("failed to get match tags: %s", err)
	}

	return services, tags, err
}

// GetConsulInfo gets the list of ryft services and
// the service tags related to requested set of files.
// the services are arranged based on "busyness" metric!
func (s *Server) getConsulInfo(userTag string, files []string) (services []*consul.CatalogService, tags []string, err error) {
	client, err := s.getConsulClient()
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
	metrics, err := s.getMetricsForAllNodes()
	if err != nil {
		return services, tags, fmt.Errorf("failed to get node metrics: %s", err)
	}
	services = s.rearrangeServices(services, metrics, s.Config.Busyness.Tolerance)
	log.WithField("metrics", metrics).Debugf("cluster node metrics")

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
		log.WithField("node", service.Node).WithField("metric", metrics[service.Node]).
			WithField("group", groupId).Debugf("service metric details")
	}

	// Go map keys are not ordered!
	// we need to sort keys by hand :(
	groupIds := make([]int, 0, len(groups))
	for groupId, _ := range groups {
		groupIds = append(groupIds, groupId)
	}
	sort.Ints(groupIds)

	// for the same group just use random shuffle
	services = make([]*consul.CatalogService, 0, len(services))
	for _, groupId := range groupIds {
		group := groups[groupId]

		// local node goes first!
		local, remote := s.splitToLocalAndRemote(group)
		if local != nil {
			services = append(services, local)
			log.WithField("node", local.Node).WithField("group", groupId).Debugf("use as local service")
		}

		// remote nodes are randomly shuffled!
		for _, k := range rand.Perm(len(remote)) {
			log.WithField("node", remote[k].Node).WithField("group", groupId).Debugf("use as remote service [%d]", k)
			services = append(services, remote[k])
		}
	}

	return services
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
	// service port must match
	if service.ServicePort != s.listenAddress.Port {
		return false
	}

	// get all interfaces
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.WithError(err).Warnf("failed to get interface addresses")
		return false
	}

	// check each interface without mask
	saddr := service.Address + "/"
	for _, addr := range addrs {
		if strings.HasPrefix(addr.String(), saddr) {
			return true
		}
	}

	return false
}

// check if service is local
func (s *Server) isLocalServiceUrl(serviceUrl *url.URL) bool {
	parts := strings.Split(serviceUrl.Host, ":")

	// service port must match
	if len(parts) < 2 || parts[1] != fmt.Sprintf("%d", s.listenAddress.Port) {
		return false
	}

	// get all interfaces
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.WithError(err).Warnf("failed to get interface addresses")
		return false
	}

	// check each interface without mask
	saddr := parts[0] + "/"
	for _, addr := range addrs {
		if strings.HasPrefix(addr.String(), saddr) {
			return true
		}
	}

	return false
}

// get service URL
func getServiceUrl(service *consul.CatalogService) string {
	scheme := "http"
	if port := service.ServicePort; port == 0 { // TODO: review the URL building!
		return fmt.Sprintf("%s://%s:8765", scheme, service.Address)
	} else {
		return fmt.Sprintf("%s://%s:%d", scheme, service.Address, port)
	}
}

// get partition info from the KV storage
// return map: mask -> list of tags
func getPartitionInfo(client *consul.Client, userTag string) (map[string][]string, error) {
	// get all wildcards (keys) and tags
	prefix := filepath.Join(userTag, "partitions") + "/"
	pairs, _, err := client.KV().List(prefix, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags from KV: %s", err)
	}

	tags := make(map[string][]string)
	for _, kvp := range pairs {
		mask, _ := url.QueryUnescape(kvp.Key)
		k := strings.TrimPrefix(mask, prefix)
		v := strings.Split(string(kvp.Value), ",")
		list := []string{}

		// trim spaces from tags, remove empty
		for _, t := range v {
			if tt := strings.TrimSpace(t); len(tt) > 0 {
				list = append(list, tt)
			}
		}

		if len(list) > 0 {
			tags[k] = list
		}
	}

	log.WithField("tags", tags).Debugf("partition info")
	return tags, nil // OK
}

// find best matched service tags for the file list
// userTag is used for multitenancy support
// no tags means "use all nodes"
func findBestMatch(client *consul.Client, userTag string, files []string) ([]string, error) {
	if len(files) == 0 {
		return nil, nil // no files - no tags
	}

	// get partition info from consul KV
	tags, err := getPartitionInfo(client, userTag)
	if err != nil {
		return nil, err
	}

	// extract keys
	keys := make([]string, 0, len(tags))
	for k, _ := range tags {
		keys = append(keys, k)
	}

	// match files and wildcards
	tags_map := make(map[string]int)
	for _, f := range files {
		// use relative path to compare, since tag keys cannot contain first '/'
		if rel, err := filepath.Rel("/", f); err == nil {
			f = rel
		}

		if found := wildmatch.IsSubsetOfAny(f, keys...); found >= 0 {
			for _, tag := range tags[keys[found]] {
				tags_map[tag] += 1
			}
		} else {
			// if no any tag found we have to search all nodes.
			// already found tags are ignored.
			log.WithField("file", f).Debugf("no any tag found for file, will search all nodes")
			return []string{}, nil // search all nodes!
		}
	}

	// map keys -> slice
	res := make([]string, 0, len(tags_map))
	for k := range tags_map {
		res = append(res, k)
	}

	return res, nil
}

// find all matched service tags for the file list
// userTag is used for multitenancy support
// no tags means "use all nodes"
func findAllMatches(client *consul.Client, userTag string, files []string) ([][]string, error) {
	if len(files) == 0 {
		return nil, nil // no files - no tags
	}

	// get partition info from consul KV
	tags, err := getPartitionInfo(client, userTag)
	if err != nil {
		return nil, err
	}

	// extract keys
	keys := make([]string, 0, len(tags))
	for k, _ := range tags {
		keys = append(keys, k)
	}

	// match files and wildcards
	res := make([][]string, len(files))
	for i, f := range files {
		tags_map := make(map[string]int)

		// use relative path to compare, since tag keys cannot contain first '/'
		if rel, err := filepath.Rel("/", f); err == nil {
			f = rel
		}

		if found := wildmatch.IsSubsetOfAny(f, keys...); found >= 0 {
			for _, tag := range tags[keys[found]] {
				tags_map[tag] += 1
			}
		} else {
			// if no any tag found we have to search all nodes.
			res[i] = nil // search all nodes!
			continue
		}

		// map keys -> slice
		for k := range tags_map {
			res[i] = append(res[i], k)
		}
	}

	return res, nil // OK
}
