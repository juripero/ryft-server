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
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used *   to endorse or promote products derived from this software without specific prior written permission. *
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
	"testing"

	"github.com/stretchr/testify/assert"
)

// parse ryftprim output
func TestParseStat(t *testing.T) {
	// check good case
	check := func(duration, totalBytes, matches int, data string) {
		s, err := ParseStat([]byte(data), "")
		if assert.NoError(t, err) && assert.NotNil(t, s) {
			assert.EqualValues(t, duration, s.Duration)
			assert.EqualValues(t, totalBytes, s.TotalBytes)
			assert.EqualValues(t, matches, s.Matches)
			// assert.InDelta(t, fabricDataRate, s.FabricDataRate)
		}
	}

	// check bad case
	bad := func(data string, expectedError string) {
		_, err := ParseStat([]byte(data), "")
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), expectedError)
		}
	}

	// bad cases
	bad(`-`, "failed to parse ryftprim output")

	bad(`
No Duration        : 1234
Total Bytes        : 1572864
Matches            : 100
Fabric Data Rate   : 12345 MB/sec
`, `failed to find "Duration" stat`)
	bad(`
Duration           : bad
Total Bytes        : 1572864
Matches            : 100
Fabric Data Rate   : 12345 MB/sec
`, `failed to parse "Duration" stat`)

	bad(`
Duration           : 1234
No Total Bytes     : 1572864
Matches            : 100
Fabric Data Rate   : 12345 MB/sec
`, `failed to find "Total Bytes" stat`)
	bad(`
Duration           : 1234
Total Bytes        : bad
Matches            : 100
Fabric Data Rate   : 12345 MB/sec
`, `failed to parse "Total Bytes" stat`)

	bad(`
Duration           : 1234
Total Bytes        : 1572864
No Matches         : 100
Fabric Data Rate   : 12345 MB/sec
`, `failed to find "Matches" stat`)
	bad(`
Duration           : 1234
Total Bytes        : 1572864
Matches            : bad
Fabric Data Rate   : 12345 MB/sec
`, `failed to parse "Matches" stat`)

	bad(`
Duration           : 1234
Total Bytes        : 1572864
Matches            : 100
No Fabric Data Rate : 12345 MB/sec
`, `failed to find "Fabric Data Rate" stat`)
	bad(`
Duration           : 1234
Total Bytes        : 1572864
Matches            : 100
Fabric Data Rate   : false
`, `failed to parse "Fabric Data Rate" stat`)
	bad(`
Duration           : 1234
Total Bytes        : 1572864
Matches            : 100
Fabric Data Rate   : bad
`, `failed to parse "Fabric Data Rate" stat from`)

	bad(`
Duration           : 1234
Total Bytes        : 1572864
Matches            : 100
Fabric Data Rate   : 12345 MB/sec
Data Rate          : false
`, `failed to parse "Data Rate" stat`)
	bad(`
Duration           : 1234
Total Bytes        : 1572864
Matches            : 100
Fabric Data Rate   : 12345 MB/sec
Data Rate          : bad
`, `failed to parse "Data Rate" stat from`)

	// bytes
	check(1234, 1572864, 100, `
 Duration           : 1234
 Total Bytes        : 1572864 bytes
 Matches            : 100
 Fabric Data Rate   : 12345 MB/sec
`)
	check(1234, 1572864, 100, `
 Duration           : 1234
 Total Bytes        : 1572864.0 bytes
 Matches            : 100
 Fabric Data Rate   : 12345 MB/sec
`)
	check(1234, 1572864, 100, `
 Duration           : 1234
 Total Bytes        : 1572864e0 bytes
 Matches            : 100
 Fabric Data Rate   : 12345 MB/sec
`)
	check(1234, 1572864, 100, `
 Duration           : 1234
 Total Bytes        : 1572864.0e0 bytes
 Matches            : 100
 Fabric Data Rate   : 12345 MB/sec
`)

	// KB
	check(1234, 1572864, 100, `
 Duration           : 1234
 Total Bytes        : 1536 KB
 Matches            : 100
 Fabric Data Rate   : 12345 MB/sec
`)
	check(1234, 1572864, 100, `
 Duration           : 1234
 Total Bytes        : 1536.0 KB
 Matches            : 100
 Fabric Data Rate   : 12345 MB/sec
`)
	check(1234, 1572864, 100, `
 Duration           : 1234
 Total Bytes        : 1536e0 KB
 Matches            : 100
 Fabric Data Rate   : 12345 MB/sec
`)
	check(1234, 1572864, 100, `
 Duration           : 1234
 Total Bytes        : 1536.0e0 KB
 Matches            : 100
 Fabric Data Rate   : 12345 MB/sec
`)

	// MB
	check(1234, 1572864, 100, `
 Duration           : 1234
 Total Bytes        : 1.5 mb
 Matches            : 100
 Fabric Data Rate   : 12345 MB/sec
`)
	check(1234, 1572864, 100, `
 Duration           : 1234
 Total Bytes        : 1.5e0 Mb
 Matches            : 100
 Fabric Data Rate   : 12345 MB/sec
`)

	_ = bad
}
