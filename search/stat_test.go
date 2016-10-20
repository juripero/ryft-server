package search

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// test empty statistics
func TestStatEmpty(t *testing.T) {
	stat := NewStat("localhost")
	assert.Empty(t, stat.Matches)
	assert.Empty(t, stat.TotalBytes)
	assert.Empty(t, stat.Duration)
	assert.Empty(t, stat.FabricDuration)
	assert.Empty(t, stat.Details)
	assert.Equal(t, "localhost", stat.Host)

	assert.Equal(t, `Stat{0 matches on 0 bytes in 0 ms (fabric: 0 ms), details:[], host:"localhost"}`, stat.String())
}

// test merge statistics (cluster mode)
func TestStatMerge(t *testing.T) {
	s1 := NewStat("")
	s1.Matches = 1
	s1.TotalBytes = 1000
	s1.Duration = 100
	s1.FabricDuration = 10
	s1.DataRate = 11.1
	s1.FabricDataRate = 111.1

	s2 := NewStat("")
	s2.Matches = 2
	s2.TotalBytes = 2000
	s2.Duration = 200
	s2.FabricDuration = 20
	s2.DataRate = 22.2
	s2.FabricDataRate = 222.2

	stat := NewStat("localhost")
	stat.Merge(nil)
	stat.Merge(s1)
	stat.Merge(s2)
	stat.Merge(nil)

	assert.Equal(t, "localhost", stat.Host)
	assert.EqualValues(t, 1+2, stat.Matches)
	assert.EqualValues(t, 1000+2000, stat.TotalBytes)
	assert.EqualValues(t, 200, stat.Duration) // maximum
	assert.InEpsilon(t, 11.1+22.2, stat.DataRate, 0.01)
	assert.EqualValues(t, 20, stat.FabricDuration) // maximum
	assert.InEpsilon(t, 111.1+222.2, stat.FabricDataRate, 0.01)
	assert.EqualValues(t, []*Stat{s1, s2}, stat.Details)

	assert.Equal(t, `Stat{3 matches on 3000 bytes in 200 ms (fabric: 20 ms), details:[Stat{1 matches on 1000 bytes in 100 ms (fabric: 10 ms), details:[], host:""} Stat{2 matches on 2000 bytes in 200 ms (fabric: 20 ms), details:[], host:""}], host:"localhost"}`, stat.String())
}

// test combine statistics (query decomposition)
func TestStatCombine(t *testing.T) {
	s1 := NewStat("")
	s1.Matches = 1
	s1.TotalBytes = 1000
	s1.Duration = 100
	s1.FabricDuration = 10
	s1.DataRate = 11.1
	s1.FabricDataRate = 111.1

	s2 := NewStat("")
	s2.Matches = 2
	s2.TotalBytes = 2000
	s2.Duration = 200
	s2.FabricDuration = 20
	s2.DataRate = 22.2
	s2.FabricDataRate = 222.2

	s3 := NewStat("")
	s3.Matches = 3
	s3.TotalBytes = 3000
	s3.Duration = 0
	s3.FabricDuration = 0
	s3.DataRate = 33.3
	s3.FabricDataRate = 333.3

	stat := NewStat("localhost")
	stat.Combine(nil)
	stat.Combine(s3)
	stat.Combine(s1)
	stat.Combine(s2)
	stat.Combine(nil)

	assert.Equal(t, "localhost", stat.Host)
	assert.EqualValues(t, 1+2+3, stat.Matches)
	assert.EqualValues(t, 1000+2000+3000, stat.TotalBytes)
	assert.EqualValues(t, 100+200, stat.Duration) // sum
	assert.InEpsilon(t, (1000+2000+3000)/1.024/1024/(100+200), stat.DataRate, 0.01)
	assert.EqualValues(t, 10+20, stat.FabricDuration) // sum
	assert.InEpsilon(t, (1000+2000+3000)/1.024/1024/(10+20), stat.FabricDataRate, 0.01)
	assert.EqualValues(t, []*Stat{s3, s1, s2}, stat.Details)

	assert.Equal(t, `Stat{6 matches on 6000 bytes in 300 ms (fabric: 30 ms), details:[Stat{3 matches on 3000 bytes in 0 ms (fabric: 0 ms), details:[], host:""} Stat{1 matches on 1000 bytes in 100 ms (fabric: 10 ms), details:[], host:""} Stat{2 matches on 2000 bytes in 200 ms (fabric: 20 ms), details:[], host:""}], host:"localhost"}`, stat.String())
}
