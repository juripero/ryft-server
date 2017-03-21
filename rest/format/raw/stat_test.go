package raw

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/getryft/ryft-server/search"
)

// compare two stats
func testStatEqual(t *testing.T, stat1, stat2 *Stat) {
	assert.EqualValues(t, stat1.Matches, stat2.Matches)
	assert.EqualValues(t, stat1.TotalBytes, stat2.TotalBytes)

	assert.EqualValues(t, stat1.Duration, stat2.Duration)
	assert.InDelta(t, stat1.DataRate, stat2.DataRate, 1.0e-3)

	assert.EqualValues(t, stat1.FabricDuration, stat2.FabricDuration)
	assert.InDelta(t, stat1.FabricDataRate, stat2.FabricDataRate, 1.0e-3)

	assert.EqualValues(t, stat1.Host, stat2.Host)
	if assert.EqualValues(t, len(stat1.Details), len(stat2.Details)) {
		for i := range stat1.Details {
			testStatEqual(t, FromStat(stat1.Details[i]), FromStat(stat2.Details[i]))
		}
	}
}

// test stat marshaling
func testStatMarshal(t *testing.T, val interface{}, expected string) {
	buf, err := json.Marshal(val)
	assert.NoError(t, err)

	assert.JSONEq(t, expected, string(buf))
}

// test STAT
func TestFormatStat(t *testing.T) {
	fmt, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, fmt)
	stat1 := fmt.NewStat()
	stat := stat1.(*Stat)
	stat.Matches = 123
	stat.TotalBytes = 456
	stat.Duration = 11
	stat.DataRate = 11.11
	stat.FabricDuration = 22
	stat.FabricDataRate = 22.22
	stat.Host = "localhost"
	// TODO: stat.Details

	stat2 := fmt.FromStat(fmt.ToStat(stat1))
	testStatEqual(t, stat1.(*Stat), stat2.(*Stat))

	testStatMarshal(t, stat1, `{"matches":123, "totalBytes":456, "duration":11, "dataRate":11.11, "fabricDuration":22, "fabricDataRate":22.22, "host":"localhost"}`)

	stat.Host = "" // should be omitted
	testStatMarshal(t, stat1, `{"matches":123, "totalBytes":456, "duration":11, "dataRate":11.11, "fabricDuration":22, "fabricDataRate":22.22}`)
}

// test raw STAT to CSV serialization
func TestStat_MarshalCSV(t *testing.T) {
	fmt, err := New()
	assert.NoError(t, err)
	assert.NotNil(t, fmt)
	stat1 := fmt.NewStat()
	stat := stat1.(*Stat)
	stat.Matches = 123
	stat.TotalBytes = 456
	stat.Duration = 11
	stat.DataRate = 11.11
	stat.FabricDuration = 22
	stat.FabricDataRate = 22.22
	stat.Host = "localhost"

	stat2 := search.NewStat("localhost")
	stat.Details = append(stat.Details , stat2)

	result, err := stat.MarshalCSV()
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"123",
		"456",
		"11",
		"11.11",
		"22",
		"22.22",
		"[{\"matches\":0,\"totalBytes\":0,\"duration\":0,\"dataRate\":0,\"fabricDuration\":0,\"fabricDataRate\":0,\"host\":\"localhost\"}]",
		"localhost",
	}, result)
}