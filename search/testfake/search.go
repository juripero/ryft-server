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

package testfake

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/getryft/ryft-server/search"
)

// Search starts asynchronous "/search" or "/count" operation.
func (engine *Engine) Search(cfg *search.Config) (*search.Result, error) {
	// report pre-defined error?
	if engine.SearchReportError != nil {
		return nil, engine.SearchReportError
	}

	// report pre-defined pseudo-random data
	if (engine.SearchReportRecords + engine.SearchReportErrors) > 0 {
		res := search.NewResult()

		go func() {
			defer res.Close()
			defer res.ReportDone()

			// report fake data
			nr := int64(engine.SearchReportRecords)
			ne := int64(engine.SearchReportErrors)
			cancelled := 0
			for (nr > 0 || ne > 0) && cancelled <= engine.SearchCancelDelay {
				if rand.Int63n(ne+nr) >= ne {
					data := []byte(fmt.Sprintf("data-%x", nr))
					idx := search.NewIndex(fmt.Sprintf("file-%d.txt", nr), uint64(nr), uint64(len(data)))
					idx.UpdateHost(engine.HostName)
					rec := search.NewRecord(idx, data)
					res.ReportRecord(rec)
					nr--
				} else {
					err := fmt.Errorf("error-%d", ne)
					res.ReportError(err)
					ne--
				}

				if res.IsCancelled() {
					cancelled++ // emulate cancel delay here
				}

				if engine.SearchReportLatency > 0 {
					time.Sleep(engine.SearchReportLatency)
				}
			}

			res.Stat = search.NewStat(engine.HostName)
			res.Stat.Matches = uint64(engine.SearchReportRecords)
			res.Stat.TotalBytes = uint64(rand.Int63n(1000000000) + 1)
			res.Stat.Duration = uint64(rand.Int63n(1000) + 1)
			res.Stat.FabricDuration = res.Stat.Duration / 2
		}()

		return res, nil // OK for now
	}

	// fake search on filesystem
	log.WithField("config", cfg).Infof("[fake]: start /search")
	engine.SearchCfgLogTrace = append(engine.SearchCfgLogTrace, *cfg)

	/*
		q, err := ryftdec.ParseQuery(cfg.Query)
		if err != nil {
			return nil, fmt.Errorf("failed to parse query: %s", err)
		}
		q = ryftdec.Optimize(q, -1)
		if q.Simple == nil {
			return nil, fmt.Errorf("no simple query parsed")
		}
	*/

	var text string
	if m := regexp.MustCompile(`\((RAW_TEXT|RECORD|RECORD\.\.*) (CONTAINS|NOT_CONTAINS|EQUALS|NOT_EQUALS) (EXACT|HAMMING|EDIT_DISTANCE)\("([^"]*)".*\)\)`).FindStringSubmatch(cfg.Query); len(m) > 4 {
		text = m[4]
	}
	if len(text) == 0 {
		return nil, fmt.Errorf("nothing to search")
	}
	log.WithField("text", text).Infof("[fake]: text to search")

	res := search.NewResult()
	go func() {
		defer log.WithField("result", res).Infof("[fake]: done /search")

		defer res.Close()
		defer res.ReportDone()

		var datWr *bufio.Writer
		var idxWr *bufio.Writer

		res.Stat = search.NewStat(engine.HostName)
		started := time.Now()
		defer func() {
			res.Stat.Duration = uint64(time.Since(started).Nanoseconds() / 1000)
			res.Stat.FabricDuration = res.Stat.Duration / 2
		}()

		for _, f := range cfg.Files {
			if res.IsCancelled() {
				break
			}

			mask := filepath.Join(engine.MountPoint, engine.HomeDir, f)
			matches, err := filepath.Glob(mask)
			//log.WithField("files", matches).WithField("mask", mask).Infof("[fake]: glob matches")
			if err != nil {
				res.ReportError(fmt.Errorf("failed to glob: %s", err))
				return // failed
			}
			for _, file := range matches {
				if res.IsCancelled() {
					break
				}

				if info, err := os.Stat(file); err != nil {
					res.ReportError(fmt.Errorf("failed to stat: %s", err))
					continue
				} else if info.IsDir() {
					continue // skip dirs
				} else if info.Size() == 0 {
					continue // skip empty files
				}

				f, err := os.Open(file)
				if err != nil {
					res.ReportError(fmt.Errorf("failed to open: %s", err))
					continue
				}
				defer f.Close()
				data, err := ioutil.ReadAll(f)
				if err != nil {
					res.ReportError(fmt.Errorf("failed to read: %s", err))
					continue
				}

				log.WithField("file", file).WithField("length", len(data)).Infof("[fake]: searching file")
				res.Stat.TotalBytes += uint64(len(data))
				for start, i := 0, 0; !res.IsCancelled() && i < 100; i++ {
					n := bytes.Index(data[start:], []byte(text))
					//log.WithField("start", start).WithField("found", n).Infof("[fake]: search iteration")
					if n < 0 {
						break
					}

					start += n + 1 // find next
					res.Stat.Matches++

					d_beg := start - 1 - int(cfg.Width)
					d_len := len(text) + 2*int(cfg.Width)
					if d_beg < 0 {
						d_len += d_beg
						d_beg = 0
					}
					if d_beg+d_len > len(data) {
						d_len = len(data) - d_beg
					}

					idx := search.NewIndex(file, uint64(d_beg), uint64(d_len))
					idx.UpdateHost(engine.HostName)
					d := data[d_beg : d_beg+d_len]
					rec := search.NewRecord(idx, d)

					if idxWr == nil && len(cfg.KeepIndexAs) != 0 {
						f, err := os.Create(filepath.Join(engine.MountPoint, engine.HomeDir, cfg.KeepIndexAs))
						if err == nil {
							log.WithField("path", f.Name()).Infof("[fake]: saving INDEX to...")
							idxWr = bufio.NewWriter(f)
							defer f.Close()
							defer idxWr.Flush()
						} else {
							log.WithError(err).Errorf("[fake]: failed to save INDEX")
						}
					}
					if datWr == nil && len(cfg.KeepDataAs) != 0 {
						f, err := os.Create(filepath.Join(engine.MountPoint, engine.HomeDir, cfg.KeepDataAs))
						if err == nil {
							log.WithField("path", f.Name()).Infof("[fake]: saving DATA to...")
							datWr = bufio.NewWriter(f)
							defer f.Close()
							defer datWr.Flush()
						} else {
							log.WithError(err).Errorf("[fake]: failed to save DATA")
						}
					}
					if idxWr != nil {
						idxWr.WriteString(fmt.Sprintf("%s,%d,%d,%d\n", idx.File, idx.Offset, idx.Length, idx.Fuzziness))
					}
					if datWr != nil {
						datWr.Write(rec.RawData)
						datWr.WriteString(cfg.Delimiter)
					}

					if cfg.ReportIndex {
						res.ReportRecord(rec)
						log.WithField("rec", rec).Infof("[fake]: report record")
					} else {
						rec.Release()
					}
				}
			}
		}
	}()

	return res, nil // OK for now

}

// drain the full results
func Drain(res *search.Result) (records []*search.Record, errors []error) {
	for {
		select {
		case err, ok := <-res.ErrorChan:
			if ok && err != nil {
				errors = append(errors, err)
			}

		case rec, ok := <-res.RecordChan:
			if ok && rec != nil {
				records = append(records, rec)
			}

		case <-res.DoneChan:
			for err := range res.ErrorChan {
				errors = append(errors, err)
			}
			for rec := range res.RecordChan {
				records = append(records, rec)
			}

			return
		}
	}
}
