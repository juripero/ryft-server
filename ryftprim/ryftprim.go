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
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/getryft/ryft-server/crpoll"
	"github.com/getryft/ryft-server/names"
	"github.com/getryft/ryft-server/records"

	"gopkg.in/yaml.v2"
)

const (
	cmd                  = "ryftprim"
	arg_type             = "-p"
	fuzzy_hamming_search = "fhs"
	arg_separator        = "-e"
	no_separator         = ""
	arg_files            = "-f"
	arg_result_file      = "-od"
	arg_index_file       = "-oi"
	arg_surrounding      = "-w"
	arg_nodes            = "-n"
	arg_case_insensetive = "-i"
	arg_fuzziness        = "-d"
	arg_query            = "-q"
	arg_verbose          = "-v"
)

const (
	ryftprimKey    = "ryftprim"
	duration       = "Duration"
	totalBytes     = "Total Bytes"
	matches        = "Matches"
	fabricDataRate = "Fabric Data Rate"
	dataRate       = "Data Rate"
)

type Params struct {
	Query         string
	Files         []string
	Surrounding   uint16
	Fuzziness     uint8
	Format        string
	CaseSensitive bool
	Nodes         uint8
	IndexFile     string
	ResultsFile   string
	KeepFiles     bool
}

type Statistics struct {
	Matches        uint32 `json:"matches"`
	TotalBytes     uint32 `json:"totalBytes"`
	Duration       uint32 `json:"duration"`
	FabricDataRate string `json:"fabricDataRate"`
	DataRate       string `json:"dataRate"`
}

func (s Statistics) AsMap() map[string]interface{} {
	m := make(map[string]interface{})
	m["duration"] = s.Duration
	m["totalBytes"] = s.TotalBytes
	m["matches"] = s.Matches
	m["fabricDataRate"] = s.FabricDataRate
	m["dataRate"] = s.DataRate
	return m
}

type Result struct {
	Errors  chan error
	Stats   chan Statistics
	Results chan records.IdxRecord
	Break   chan struct{}
}

func newResult() *Result {
	return &Result{
		make(chan error, 1),
		make(chan Statistics, 1),
		make(chan records.IdxRecord, 256),
		make(chan struct{}, 1),
	}
}

func uintOrPanic(str string) uint32 {
	if val, err := strconv.ParseUint(str, 10, 32); err != nil {
		panic(err)
	} else {
		return uint32(val)
	}
}

func StatisticsFromMap(m map[string]string) Statistics {
	return Statistics{
		uintOrPanic(m[matches]),
		uintOrPanic(m[totalBytes]),
		uintOrPanic(m[duration]),
		m[fabricDataRate],
		m[dataRate],
	}
}

//type RyftprimStats map[string]interface{}

func Search(p *Params) (result *Result) {
	result = newResult()

	// prepare command line arguments
	testArgs := []string{
		arg_type, fuzzy_hamming_search,
		arg_separator, no_separator,
		arg_verbose,
	}

	if !p.CaseSensitive {
		testArgs = append(testArgs, arg_case_insensetive)
	}

	if p.IndexFile != "" {
		testArgs = append(testArgs, arg_index_file, p.IndexFile)
	}

	if p.ResultsFile != "" {
		testArgs = append(testArgs, arg_result_file, p.ResultsFile)
	}

	for _, file := range p.Files {
		testArgs = append(testArgs, arg_files, file)
	}

	if p.Nodes > 0 {
		testArgs = append(testArgs, arg_nodes, fmt.Sprintf("%d", p.Nodes))
	}

	if p.Surrounding > 0 {
		testArgs = append(testArgs, arg_surrounding, fmt.Sprintf("%d", p.Surrounding))
	}

	if p.Fuzziness > 0 {
		testArgs = append(testArgs, arg_fuzziness, fmt.Sprintf("%d", p.Fuzziness))
	}

	testArgs = append(testArgs, arg_query, p.Query)

	log.Println(testArgs)

	go func() {
		// execute ryftprim and get output
		command := exec.Command(cmd, testArgs...)
		output, err := command.CombinedOutput()

		defer close(result.Results)
		defer close(result.Stats)
		defer close(result.Errors)

		outputstr := string(output)
		log.Printf("\n%s\n", outputstr)
		result.Errors <- nil // done

		if err != nil {
			result.Errors <- errors.New(fmt.Sprintf("%s (%s)", strings.TrimSpace(outputstr), err.Error()))
			return
		}

		// parse output results for statistics numbers
		statsmap := make(map[string]string)
		err = yaml.Unmarshal(output, statsmap)

		if err != nil {
			result.Errors <- errors.New(fmt.Sprintf("%s (%s)", strings.TrimSpace(outputstr), err.Error()))
			return
		}

		result.Stats <- StatisticsFromMap(statsmap)
	}()

	// if we have index then let's read results
	if p.IndexFile != "" {
		go func() {
			// read an index file
			var idx, res *os.File
			var err error
			if idx, err = crpoll.OpenFile(names.ResultsDirPath(p.IndexFile), result.Errors); err != nil {
				result.Errors <- err
				return
			}
			defer cleanup(idx, p.KeepFiles)

			//read a results file
			if res, err = crpoll.OpenFile(names.ResultsDirPath(p.ResultsFile), result.Errors); err != nil {
				result.Errors <- err
				return
			}
			defer cleanup(res, p.KeepFiles)

			indexes, drop := records.Poll(idx, result.Errors)
			recs := records.DataPoll(indexes, res)

			for {
				select {
				case item, rok := <-recs:
					if !rok {
						return
					}
					result.Results <- item
				case d, dok := <-result.Break:
					if !dok {
						return
					}
					drop <- d
				}
			}

		}()
	}

	return
}

func cleanup(file *os.File, keep bool) {
	if file != nil {
		log.Printf(" Close file %v", file.Name())
		file.Close()
		if !keep {
			os.Remove(file.Name())
		}
	}
}

func createRyftprimStatistic(m map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"duration":       m[duration],
		"totalBytes":     m[totalBytes],
		"matches":        m[matches],
		"fabricDataRate": m[fabricDataRate],
		"dataRate":       m[dataRate],
	}
}
