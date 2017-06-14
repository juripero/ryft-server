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
	"strings"
	"time"
)

// add new pending job
func (server *Server) addJob(cmd, args string, when time.Time) {
	jobsLog.WithFields(map[string]interface{}{
		"cmd":  cmd,
		"args": args,
		"when": when,
	}).Debugf("[%s]: adding new job", JOBS)

	server.settings.AddJob(cmd, args, when)
	// atomic.AddInt32(&server.newJobsCount, 1)
	server.gotJobsChan <- 1 // notify processing goroutine about new job
	// TODO: do not notify many times
}

// start new processing goroutine
func (server *Server) startJobsProcessing() {
	server.gotJobsChan = make(chan int, 256)
	go server.processJobs()
}

// process pending jobs (separate goroutine)
func (server *Server) processJobs() {
	defer func() {
		if r := recover(); r != nil {
			log.WithField("error", r).Errorf("[%s]: process jobs failed", CORE)
		}
	}()

	// sleep a while before start
	time.Sleep(1 * time.Second)

	for {
		now := time.Now()

		// get Job list to be done (1 second in advance)
		jobsLog.WithField("time", now).Debugf("[%s]: getting pending jobs...", JOBS)
		jobs, err := server.settings.QueryAllJobs(now.Add(1 * time.Second))
		if err != nil {
			jobsLog.WithError(err).Warnf("[%s]: failed to get pending jobs", JOBS)
			time.Sleep(10 * time.Second)
		}

		// do jobs
		ids := []int64{} // completed
		for job := range jobs {
			if server.doJob(job) {
				ids = append(ids, job.Id)
			}
		}

		// delete completed jobs
		if len(ids) > 0 {
			jobsLog.WithField("jobs", ids).Debugf("[%s]: jobs are completed, deleting...", JOBS)
			if err = server.settings.DeleteJobs(ids); err != nil {
				log.WithError(err).Warnf("[%s]: failed to delete completed jobs", JOBS)
			}
		}

		next, err := server.settings.GetNextJobTime()
		if err != nil {
			jobsLog.WithError(err).Warnf("[%s]: failed to get next job time", JOBS)
			next = now.Add(1 * time.Hour)
		}
		jobsLog.WithField("time", next).Debugf("[%s]: next job time", JOBS)

		sleep := next.Sub(now)
		if sleep < time.Second {
			sleep = time.Second
		}

		jobsLog.WithField("sleep", sleep).Debugf("[%s]: sleep a while before next iteration...", JOBS)
		select {
		case <-time.After(sleep):
			continue

		case <-server.gotJobsChan:
			// atomic.AddInt32(&server.newJobsCount, -1)
			continue

		case <-server.closeCh:
			return
		}
	}
}

// do pending job
func (server *Server) doJob(job SettingsJobItem) bool {
	switch strings.ToLower(job.Cmd) {
	case "delete-file":
		res := deleteAll("/", []string{job.Args})
		jobsLog.WithFields(map[string]interface{}{
			"file":   job.Args,
			"result": res,
		}).Debugf("[%s]: delete file", JOBS)
		return true

	case "delete-catalog":
		res := deleteAll("/", []string{job.Args})
		jobsLog.WithFields(map[string]interface{}{
			"catalog": job.Args,
			"result":  res,
		}).Debugf("[%s]: delete catalog", JOBS)
		return true
	}

	jobsLog.WithFields(map[string]interface{}{
		"cmd":  job.Cmd,
		"args": job.Args,
	}).Warnf("[%s]: unknown command, ignored", JOBS)
	// return false // will be processed later
	return true // ignore job
}
