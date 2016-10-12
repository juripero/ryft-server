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
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
)

// update busyness thread
func (s *Server) startUpdatingBusyness() {
	s.busynessChanged = make(chan int32, 256)

	// TODO: sleep a while before start?

	// dedicated goroutine to monitor and update metric
	go func(metric int32) {
		var reported int32 = -1 // to force update metric ASAP

		for {
			select {
			case metric = <-s.busynessChanged:
				busyLog.WithField("metric", metric).Debugf("[%s]: metric changed", BUSY)
				continue

			case <-time.After(s.Config.Busyness.UpdateLatency):
				if metric != reported {
					reported = metric
					busyLog.WithField("metric", metric).Debugf("[%s]: metric reporting...", BUSY)
					if err := s.updateConsulMetric(int(metric)); err != nil {
						busyLog.WithError(err).Warnf("[%s]: failed to update metric", BUSY)
					}
				}

				// TODO: graceful goroutine shutdown
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
