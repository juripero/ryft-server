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
			// assert.InEpsilon(t, fabricDataRate, s.FabricDataRate)
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
