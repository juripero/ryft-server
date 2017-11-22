package datetime

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// LoadTimezone with the support of UTC offsets
func LoadTimezone(v string) (*time.Location, error) {
	if v == "" {
		return time.UTC, nil
	}
	loc, err := time.LoadLocation(v)
	if err == nil {
		return loc, nil
	}
	offset, err := parseUTCOffset(v)
	if err != nil {
		return nil, fmt.Errorf(`failed to parse timezone %s with error: %s`, v, err)
	}
	loc = time.FixedZone(time.UTC.String(), offset)
	return loc, nil
}

// get offset in seconds for strings like
// -01:00; 08:00; -01; +08; -0100; 0800
func parseUTCOffset(v string) (int, error) {
	// detect negative
	var (
		result, h, m int
		err          error
		neg          = false
	)
	p := v[0]
	if p == '-' || p == '+' {
		neg = p == '-'
		v = v[1:]
	}

	if strings.HasPrefix(v, ":") {
		return 0, errors.New(`offset has an unexpected format`)
	}
	v = strings.Replace(v, ":", "", 1)
	vSize := len(v)
	if vSize == 2 || vSize == 4 {
		h, err = strconv.Atoi(v[:2])
		if err != nil {
			return 0, errors.New(`failed to parse offset hours`)
		}
		if vSize == 4 {
			m, err = strconv.Atoi(v[2:4])
			if err != nil {
				return 0, errors.New(`failed to parse offset minutes`)
			}
		}
	} else {
		return 0, errors.New(`offset has an unexpected format`)
	}
	result = (h*60 + m) * 60
	if neg {
		result *= -1
	}

	return result, nil
}
