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

package ryftprim

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/mitchellh/mapstructure"
)

// dedicated structure for aggregation goroutine
type aggregationGoroutine struct {
	aggregations search.Aggregations
	lastError    error
}

// AggregationOptions contains static and runtime aggregation options
type AggregationOptions struct {
	ToolPath []string `json:"optimized-tool,omitempty"`

	// runtime options
	Engine             string `json:"engine,omitempty"`
	MaxRecordsPerChunk string `json:"max-records-per-chunk,omitempty"`
	DataChunkSize      string `json:"data-chunk-size,omitempty"`
	IndexChunkSize     string `json:"index-chunk-size,omitempty"`
	Concurrency        int    `json:"concurrency,omitempty"`
}

// inverse of Parse() method
func (opts *AggregationOptions) ToMap() map[string]interface{} {
	// do it via JSON encoding/decoding
	data, _ := json.Marshal(opts)
	var res map[string]interface{}
	_ = json.Unmarshal(data, &res)
	return res
}

// Parse aggregation options
func (opts *AggregationOptions) Parse(params interface{}, keepToolPath bool) error {
	var oldToolPath []string
	if keepToolPath {
		oldToolPath = opts.ToolPath
		opts.ToolPath = nil // reset
	}

	dcfg := mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           opts,
		TagName:          "json",
	}
	if d, err := mapstructure.NewDecoder(&dcfg); err != nil {
		return fmt.Errorf("failed to create decoder: %s", err)
	} else if err := d.Decode(params); err != nil {
		return fmt.Errorf("failed to decode: %s", err)
	}

	// path to optimized tool
	if keepToolPath {
		// tool-path cannot be changed by user options
		opts.ToolPath = oldToolPath
	} else if len(opts.ToolPath) != 0 {
		if info, err := os.Stat(opts.ToolPath[0]); os.IsNotExist(err) {
			return fmt.Errorf(`no "optimized-tool" tool found: %s`, err)
		} else if err != nil {
			return fmt.Errorf(`bad "optimized-tool" tool found: %s`, err)
		} else if info.IsDir() {
			return fmt.Errorf(`bad "optimized-tool" tool found: %s`, "is a directory")
		}
	}

	// concurrency
	if opts.Concurrency < 0 {
		return fmt.Errorf(`"concurrency" cannot be negative`)
	}

	return nil // OK
}

// Apply aggregations
func ApplyAggregations(opts AggregationOptions, indexPath, dataPath string, delimiter string,
	aggregations search.Aggregations, checkJsonArray bool, cancelFunc func() bool) error {

	if len(opts.ToolPath) != 0 {
		// TODO: use optimized tool to calculate aggregations
	}

	var idxRd, datRd *bufio.Reader
	var dataPos uint64 // DATA read position
	var dataSkip uint64

	// open INDEX file
	if idxRd == nil {
		f, err := os.Open(indexPath)
		if err != nil {
			log.WithError(err).WithField("path", indexPath).
				Warnf("[%s/aggs]: failed to open INDEX file", TAG)
			return fmt.Errorf("failed to open INDEX file: %s", err)
		}

		defer f.Close() // close at the end
		idxRd = bufio.NewReaderSize(f, ReadBufSize)
	}

	// open DATA file
	if datRd == nil {
		f, err := os.Open(dataPath)
		if err != nil {
			log.WithError(err).WithField("path", dataPath).
				Warnf("[%s/aggs]: failed to open DATA file", TAG)
			return fmt.Errorf("failed to open DATA file: %s", err)
		}

		defer f.Close() // close at the end
		datRd = bufio.NewReaderSize(f, ReadBufSize)
		if checkJsonArray {
			if jarr, err := IsJsonArray(datRd); err != nil {
				return fmt.Errorf("failed to check JSON array: %s", err)
			} else if jarr {
				dataSkip = JsonArraySkip // JSON array marker
			}
		}
	}

	var subAggs []*aggregationGoroutine
	var dataCh chan []byte
	var wg sync.WaitGroup
	var subErrs int32
	if opts.Concurrency > 1 {
		// create a few goroutines to process aggregations
		// each goroutine will use its own Aggerations
		subAggs = make([]*aggregationGoroutine, opts.Concurrency)
		dataCh = make(chan []byte, 4*1024)
		log.Debugf("[%s/aggs]: start sub-processing in %d threads", TAG, opts.Concurrency)
		start := time.Now()

		// run several processing goroutines
		for i := range subAggs {
			subAggs[i] = &aggregationGoroutine{
				aggregations: aggregations.Clone(),
			}

			wg.Add(1)
			go func(a *aggregationGoroutine) {
				defer func() {
					if r := recover(); r != nil {
						log.Warnf("UNHANDLED PANIC: %s", r)
					}
				}()
				defer wg.Done()

				for {
					if data, ok := <-dataCh; ok {
						// apply aggregations
						if err := a.aggregations.Add(data); err != nil {
							a.lastError = fmt.Errorf("failed to apply aggregation: %s", err)
							atomic.AddInt32(&subErrs, 1) // notify main thread
							return
						}
					} else {
						break // done
					}
				}
			}(subAggs[i])
		}

		// at the end combine all sub-aggregations
		defer func() {
			log.Debugf("[%s/aggs]: waiting sub-processing", TAG)
			close(dataCh)
			wg.Wait()
			log.Debugf("[%s/aggs]: sub-processing done in %s", TAG, time.Since(start))

			// print errors to the log
			for _, sa := range subAggs {
				if err := sa.lastError; err != nil {
					log.WithError(err).Warnf("[%s/aggs]: sub-processing error", TAG)
				}

				if err := aggregations.Merge(sa.aggregations.ToJson(false)); err != nil {
					log.WithError(err).Warnf("[%s/aggs]: sub-process merge error", TAG)
				}
			}
		}()
	}

	// read INDEX line-by-line and corresponding DATA
	for {
		line, err := idxRd.ReadBytes('\n')

		if err != nil {
			if err == io.EOF {
				if len(line) == 0 {
					return nil // DONE
				}
				// ... process line even err == EOF
			} else {
				log.WithError(err).Warnf("[%s/aggs]: INDEX reading failed", TAG)
				return fmt.Errorf("INDEX reading failed: %s", err)
			}
		}

		//		log.WithField("line", string(bytes.TrimSpace(line))).
		//			Debugf("[%s/aggs]: new INDEX line read", TAG) // FIXME: DEBUG

		index, err := search.ParseIndex(line)
		if err != nil {
			log.WithError(err).Warnf("[%s/aggs]: failed to parse INDEX from %q", TAG, bytes.TrimSpace(line))
			return fmt.Errorf("failed to parse INDEX: %s", err)
		}

		// skip JSON array mark
		if n := int(dataSkip); n != 0 {
			m, err := datRd.Discard(n)
			if err != nil {
				log.WithError(err).Warnf("[%s/aggs]: failed to skip JSON mark", TAG)
				return fmt.Errorf("failed to skip JSON mark: %s", err)
			} else if m != n {
				log.Warnf("[%s/aggs]: not all JSON mark skipped: %d of %d", TAG, m, n)
				return fmt.Errorf("not all JSON mark skipped: %d of %d", m, n)
			}
		}

		data := make([]byte, index.Length)
		if n, err := io.ReadFull(datRd, data); err != nil {
			log.WithError(err).Warnf("[%s/aggs]: failed to read DATA", TAG)
			return fmt.Errorf("failed to read DATA: %s", err)
		} else if uint64(n) != index.Length {
			log.Warnf("[%s/aggs]: not all DATA read: %d of %d", TAG, n, index.Length)
			return fmt.Errorf("not all DATA read: %d of %d", n, index.Length)
		}
		dataPos += dataSkip + uint64(len(data))

		//		log.WithField("data", string(data)).
		//			Debugf("[%s/aggs]: new DATA read", TAG) // FIXME: DEBUG

		// read and check delimiter
		if len(delimiter) > 0 {
			if n, err := datRd.Discard(len(delimiter)); err != nil {
				log.WithError(err).Warnf("[%s/aggs]: failed to discard DATA delimiter", TAG)
				return fmt.Errorf("failed to discard DATA delimiter: %s", err)
			} else if n != len(delimiter) {
				log.Warnf("[%s/aggs]: not all DATA delimiter discarded: %d of %d", TAG, n, len(delimiter))
				return fmt.Errorf("not all DATA deliiter discarded: %d of %d", n, len(delimiter))
			}

			dataPos += uint64(len(delimiter))
		}

		if opts.Concurrency > 1 {
			dataCh <- data // send data to processing goroutines
			if atomic.LoadInt32(&subErrs) != 0 {
				return fmt.Errorf("parallel error occurred")
			}
		} else {
			// apply aggregations
			if err := aggregations.Add(data); err != nil {
				return fmt.Errorf("failed to apply aggregation: %s", err)
			}
		}

		index.Release()

		// check "cancel" channel
		if cancelFunc != nil && cancelFunc() {
			log.Debugf("[%s/aggs]: cancelled", TAG)
			return nil // cancelled
		}
	}
}
