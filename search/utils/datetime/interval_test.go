package datetime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInterval(t *testing.T) {
	assert := assert.New(t)
	date := time.Date(2009, 4, 2, 3, 4, 5, 6, time.UTC)
	check := func(i string, expected time.Time) {
		proc := NewInterval(i)
		err := proc.Parse()
		assert.NoError(err)
		date := proc.Apply(date)
		assert.WithinDuration(expected, date, time.Millisecond)
	}
	check("year", time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC))
	check("1y", time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC))
	check("month", time.Date(2009, 4, 1, 0, 0, 0, 0, time.UTC))
	check("1M", time.Date(2009, 4, 1, 0, 0, 0, 0, time.UTC))
	check("quarter", time.Date(2009, 1, 16, 0, 0, 0, 0, time.UTC))
}
