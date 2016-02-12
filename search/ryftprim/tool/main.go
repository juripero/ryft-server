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
	"encoding/json"
	"fmt"
	"os"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftone"
)

// application configuration
type Config struct {
	RawCfg    search.Config `json:"rawCfg"`
	IndexFile string        `json:"indexFile"`
	DataFile  string        `json:"dataFile"`
}

// just print expected configuration
func main0() {
	var cfg Config
	cfg.DataFile = "data.bin"
	cfg.IndexFile = "index.txt"

	cfg.RawCfg.Query = "555"
	cfg.RawCfg.AddFile("/regression/*.txt")
	cfg.RawCfg.Surrounding = 11
	cfg.RawCfg.Fuzziness = 2
	cfg.RawCfg.CaseSensitive = false
	cfg.RawCfg.Nodes = 3

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(cfg)
}

// application entry point
func main() {
	var cfg Config
	dec := json.NewDecoder(os.Stdin)
	if err := dec.Decode(&cfg); err != nil {
		fmt.Printf("failed to decode search config: %s\n", err)
		os.Exit(-1)
	}
	raw := cfg.RawCfg

	// create data set
	ds, err := ryftone.NewDataSet(raw.Nodes)
	if err != nil {
		fmt.Printf("failed to create dataset: %s\n", err)
		os.Exit(-2)
	}
	defer ds.Delete()

	// files
	for _, file := range raw.Files {
		err = ds.AddFile(file)
		if err != nil {
			fmt.Printf("failed to add %q file: %s\n", file, err)
			os.Exit(-3)
		}
	}

	// run search
	err = ds.SearchFuzzyHamming(ryftone.DetectPlainQuery(raw.Query),
		cfg.DataFile, cfg.IndexFile, raw.Surrounding,
		raw.Fuzziness, raw.CaseSensitive)
	if err != nil {
		fmt.Printf("failed to run search: %s\n", err)
		os.Exit(-4)
	}

	matches := ds.GetTotalMatches()
	totalBytes := ds.GetTotalBytesProcessed()
	duration := ds.GetExecutionDuration()
	dataRate := ryftone.BpmsToMbps(totalBytes, duration)
	fabricDuration := ds.GetFabricExecutionDuration()
	fabricDataRate := ryftone.BpmsToMbps(totalBytes, fabricDuration)

	fmt.Printf("Matches: %d\n", matches)
	fmt.Printf("Total Bytes: %d\n", totalBytes)
	fmt.Printf("Duration: %d\n", duration)
	fmt.Printf("Fabric Duration: %d\n", fabricDuration)
	fmt.Printf("Fabric Data Rate: %f MB/sec\n", fabricDataRate)
	fmt.Printf("Data Rate: %f MB/sec\n", dataRate)
}
