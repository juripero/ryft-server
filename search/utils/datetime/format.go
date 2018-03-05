package datetime

import (
	"time"

	"github.com/araddon/dateparse"
	"github.com/vjeantet/jodaTime"
)

// ISO8601toGo convert ISO-8601 (joda) format to Go date format (e.g. "yyyy-MM-dd" -> "2006-04-02")
func FormatAsISO8601(joda string, t time.Time) string {
	return jodaTime.Format(joda, t)
}

// ParseIn parse string and returns time.Time object in selected timezone
func ParseIn(s string, timezone *time.Location) (time.Time, error) {
	return dateparse.ParseIn(s, timezone)
}
