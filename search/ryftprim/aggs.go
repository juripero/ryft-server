/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2018, Ryft Systems, Inc.
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
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils/aggs"
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

// ParseConfig parses aggregation options from configuration file
func (opts *AggregationOptions) ParseConfig(params interface{}) error {
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
	if len(opts.ToolPath) != 0 {
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

// ParseTweaks parses aggregation options tweaks (user customized)
func (opts *AggregationOptions) ParseTweaks(params interface{}) error {
	// ToolPath should not be changed via tweaks!
	oldToolPath := opts.ToolPath
	opts.ToolPath = nil // reset

	if err := opts.ParseConfig(params); err != nil {
		return err
	}

	// restore ToolPath
	opts.ToolPath = oldToolPath
	return nil // OK
}

// Apply optimized (c-based) aggregations
/*
Use the following command to check:
curl -s "http://localhost:8765/search/aggs?data=test-10M.bin&index=test-10M.txt&delimiter=%0d%0a&format=json&performance=true" --data '{"tweaks":{"aggs":{"concurrency":4,"engine":"optimized"}}, "aggs":{"1":{"stats":{"field":"foo"}},"2":{"avg":{"field":"foo"}}}}' | jq .stats.extra
*/
func applyOptimizedAggregations(opts AggregationOptions, indexPath, dataPath string, delimiter string,
	aggregations search.Aggregations, checkJsonArray bool, cancelFunc func() bool) (error, bool) {

	log.WithField("opts", opts.ToMap()).Debugf("[%s/aggs]: try to apply optimized version", TAG)

	// check optimized tool is configured
	if len(opts.ToolPath) == 0 {
		return nil, false // no optimized tool configured
	}

	// check the "engine"
	forceOptimized := false
	switch strings.ToLower(opts.Engine) {
	case "native":
		return nil, false // disabled by user

	case "optimized":
		forceOptimized = true
		break

	case "auto", "":
		break

	default:
		return fmt.Errorf(`"%s" is unknown aggregation engine (should be "native", "optimized" or "auto")`, opts.Engine), true
	}

	if a, ok := aggregations.(*aggs.Aggregations); ok {
		// check the "json" format
		switch a.Format {
		case "json":
			break

		default:
			if forceOptimized {
				return fmt.Errorf("requested aggregation format is not supported by optimized engine"), true
			}
			return nil, false // not a "json" format found
		}

		// check we have only "stat" aggregations
		for _, e := range a.Engines {
			if _, ok := e.(*aggs.Stat); !ok {
				if forceOptimized {
					return fmt.Errorf("requested type of aggregation is not supported by optimized engine"), true
				}
				return nil, false // not a "stat" engine found
			}
		}

		// prepare command line arguments
		args := make([]string, 0, len(opts.ToolPath)+16)
		args = append(args, opts.ToolPath...)
		if opts.DataChunkSize != "" {
			args = append(args, "--data-chunk", opts.DataChunkSize)
		}
		if opts.IndexChunkSize != "" {
			args = append(args, "--index-chunk", opts.IndexChunkSize)
		}
		if opts.MaxRecordsPerChunk != "" {
			args = append(args, "--max-records", opts.MaxRecordsPerChunk)
		}
		args = append(args, fmt.Sprintf("--concurrency=%d", opts.Concurrency))
		args = append(args, fmt.Sprintf("--delimiter=%d", len(delimiter)))
		if checkJsonArray {
			if jarr, err := IsJsonArrayFile(dataPath); err != nil {
				return fmt.Errorf("failed to check JSON array: %s", err), true
			} else if jarr {
				args = append(args,
					fmt.Sprintf("--delimiter=%d", len(delimiter)+JsonArraySkip), // override
					fmt.Sprintf("--header=%d", JsonArraySkip),
					fmt.Sprintf("--footer=%d", JsonArraySkip))
			}
		}
		args = append(args, "--data", dataPath, "--index", indexPath)
		engines := make([]aggs.Engine, 0, len(a.Engines)) // engines in fixed order
		for _, e := range a.Engines {
			if s, ok := e.(*aggs.Stat); ok {
				args = append(args, "--field", s.Field.String())
				engines = append(engines, e)
			}
		}
		args = append(args, "-q", "--native")

		// run optimized engine on dedicated goroutine
		log.WithField("args", args).Debugf("[%s/aggs]: use 'optimized' engine", TAG)
		cmd := exec.Command(args[0], args[1:]...)
		cmdCh := make(chan struct{})
		var cmdOut []byte
		var cmdErr error
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("UNHANDLED PANIC: %s", r)
					cmdErr = fmt.Errorf("unhandled panic: %s", r)
				}

				close(cmdCh)
			}()

			cmdOut, cmdErr = cmd.CombinedOutput()
		}()

		// busy-wait the optimized engine is done
	BusyWait:
		for {
			select {
			case <-cmdCh:
				// tool is done
				break BusyWait

			case <-time.After(100 * time.Millisecond):
				if cancelFunc() {
					// check if request is cancelled
					if p := cmd.Process; p != nil {
						if err := p.Kill(); err != nil {
							log.WithError(err).Warnf("[%s/aggs]: failed to kill 'optimized' engine", TAG)
						}
					}
				}
			}
		}
		if cmdErr != nil {
			return fmt.Errorf("failed to run optimized engine: %s\n%s", cmdErr, cmdOut), true
		}

		// parse results
		var jdata []interface{}
		if err := json.Unmarshal(cmdOut, &jdata); err != nil {
			return fmt.Errorf("failed to decode optimized engine response: %s", err), true
		} else if len(jdata) != len(a.Engines) {
			return fmt.Errorf("bad optimized engine response: length mismath (%d != %d)", len(jdata), len(a.Engines)), true
		}

		// merge resutls
		for i, e := range engines {
			if err := e.Merge(jdata[i]); err != nil {
				return fmt.Errorf("failed to merge optimized engine response: %s", err), true
			}
		}

		return nil, true // OK
	}

	return nil, false // bad aggregation type
}

// Apply aggregations
func ApplyAggregations(opts AggregationOptions, indexPath, dataPath string, delimiter string,
	aggregations search.Aggregations, checkJsonArray bool, cancelFunc func() bool) error {

	// try to use optimized tool to calculate aggregations
	if err, ok := applyOptimizedAggregations(opts, indexPath, dataPath,
		delimiter, aggregations, checkJsonArray, cancelFunc); ok {
		return err
	}

	log.Debugf("[%s/aggs]: use 'native' engine", TAG)

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
