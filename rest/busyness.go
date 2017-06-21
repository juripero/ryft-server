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
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
	consul "github.com/hashicorp/consul/api"
)

// update busyness thread
func (server *Server) startUpdatingBusyness() {
	server.busynessChanged = make(chan int32, 256)

	// TODO: sleep a while before start?

	// dedicated goroutine to monitor and update metric
	go func(metric int32) {
		var reported int32 = -1 // to force update metric ASAP

		defer func() {
			if r := recover(); r != nil {
				log.WithField("error", r).Errorf("[%s]: update busyness thread failed", CORE)
			}
		}()

		for {
			select {
			case metric = <-server.busynessChanged:
				busyLog.WithField("metric", metric).Debugf("[%s]: metric changed", BUSY)
				continue

			case <-time.After(server.Config.Busyness.UpdateLatency):
				if metric != reported {
					reported = metric
					busyLog.WithField("metric", metric).Debugf("[%s]: metric reporting...", BUSY)
					if err := server.updateNodeMetric(int(metric)); err != nil {
						busyLog.WithError(err).Warnf("[%s]: failed to update metric", BUSY)
					}
				}

			case <-server.closeCh:
				return
			}
		}
	}(server.activeSearchCount)
}

// notify server a search is started
func (server *Server) onSearchStarted(config *search.Config) {
	server.onSearchChanged(config, +1)
}

// notify server a search is started
func (server *Server) onSearchStopped(config *search.Config) {
	server.onSearchChanged(config, -1)
}

// notify server a search is changed
func (server *Server) onSearchChanged(config *search.Config, delta int32) {
	metric := atomic.AddInt32(&server.activeSearchCount, delta)
	if server.busynessChanged != nil {
		// notify to update metric
		server.busynessChanged <- metric
	}
}

// update the node metric in the cluster
func (server *Server) updateNodeMetric(metric int) error {
	client, err := server.getConsulClient()
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
func (server *Server) getMetricsForAllNodes() (map[string]int, error) {
	client, err := server.getConsulClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get consul client: %s", err)
	}

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

	return metrics, nil // OK
}
