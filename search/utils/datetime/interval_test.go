package datetime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIntervalAlignment(t *testing.T) {
	assert := assert.New(t)
	date := time.Date(2009, 4, 2, 3, 4, 5, 6, time.UTC)
	check := func(i string, expected time.Time) {
		proc := NewInterval(i)
		err := proc.Parse()
		assert.NoError(err)
		date := proc.Truncate(date)
		assert.WithinDuration(expected, date, time.Millisecond)
	}
	check("year", time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC))
	check("1y", time.Date(2009, 1, 1, 0, 0, 0, 0, time.UTC))
	check("month", time.Date(2009, 4, 1, 0, 0, 0, 0, time.UTC))
	check("1M", time.Date(2009, 4, 1, 0, 0, 0, 0, time.UTC))
	check("2M", time.Date(2009, 3, 1, 0, 0, 0, 0, time.UTC))
	check("quarter", time.Date(2009, 4, 1, 0, 0, 0, 0, time.UTC))
	check("week", time.Date(2009, 3, 29, 0, 0, 0, 0, time.UTC))
	check("1w", time.Date(2009, 3, 29, 0, 0, 0, 0, time.UTC))
	check("2w", time.Date(2009, 3, 29, 0, 0, 0, 0, time.UTC))
	check("day", time.Date(2009, 4, 2, 0, 0, 0, 0, time.UTC))
	check("2d", time.Date(2009, 4, 2, 0, 0, 0, 0, time.UTC))
	check("100d", time.Date(2008, 12, 25, 0, 0, 0, 0, time.UTC))
}
