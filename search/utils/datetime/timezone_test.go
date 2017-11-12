package datetime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadTimezone(t *testing.T) {
	assert := assert.New(t)
	success := func(name string, expected string) {
		loc, err := LoadTimezone(name)
		assert.NoError(err)
		moment := time.Date(2007, 1, 1, 12, 0, 0, 0, time.UTC)
		assert.Equal(expected, moment.In(loc).Format("Z07:00"))
	}

	fail := func(name string) {
		_, err := LoadTimezone(name)
		assert.Error(err)
	}

	success("", "Z")
	success("UTC", "Z")
	success("America/Chicago", "-06:00")
	success("Asia/Pontianak", "+07:00")
	success("08:00", "+08:00")
	success("-08:00", "-08:00")
	success("+01", "+01:00")
	success("-01", "-01:00")
	success("0200", "+02:00")
	success("-0200", "-02:00")
	success("+10:30", "+10:30")
	success("+04:51", "+04:51")

	fail("Russia/Cheboksary")
	fail("1")
	fail("+1")
	fail("-1")
	fail("-x1")
	fail("-:")
	fail("-:0")
	fail("-:01")
}
