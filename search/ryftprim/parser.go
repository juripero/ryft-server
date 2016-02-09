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
	"fmt"

	"gopkg.in/yaml.v2"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/ryftone"
	"github.com/getryft/ryft-server/search/utils"
)

// parseIndex parses Index record from custom line.
func parseIndex(buf []byte) (index search.Index, err error) {
	return ryftone.ParseIndex(buf)
}

// parseStat parses statistics from ryftprim output.
func parseStat(buf []byte) (stat *search.Statistics, err error) {
	// parse as YML map first
	v := map[string]interface{}{}
	err = yaml.Unmarshal(buf, &v)
	if err != nil {
		return stat, fmt.Errorf("failed to parse ryftprim output: %s", err)
	}

	log.WithField("stat", v).Debugf("[%s] output as YML", TAG)
	stat = search.NewStat()

	// Duration
	stat.Duration, err = utils.AsUint64(v["Duration"])
	if err != nil {
		return nil, fmt.Errorf(`failed to parse "Duration" stat`)
	}

	// Total Bytes
	stat.TotalBytes, err = utils.AsUint64(v["Total Bytes"])
	if err != nil {
		return nil, fmt.Errorf(`failed to parse "Total Bytes" stat`)
	}

	// Matches
	stat.Matches, err = utils.AsUint64(v["Matches"])
	if err != nil {
		return nil, fmt.Errorf(`failed to parse "Matches" stat`)
	}

	// Fabric Data Rate
	_, err = utils.AsString(v["Fabric Data Rate"])
	if err != nil {
		return nil, fmt.Errorf(`failed to parse "Fabric Data Rate" stat`)
	}

	// Data Rate
	_, err = utils.AsString(v["Data Rate"])
	if err != nil {
		return nil, fmt.Errorf(`failed to parse "Data Rate" stat`)
	}

	return stat, nil // OK
}
