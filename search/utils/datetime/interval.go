package datetime

import (
	"fmt"
	"go/scanner"
	"go/token"
	"math"
	"strconv"
	"time"
)

const (
	UNIT_YEAR = iota
	UNIT_QUARTER
	UNIT_MONTH
	UNIT_WEEK
	UNIT_DAY
	UNIT_HOUR
	UNIT_MINUTE
	UNIT_SECOND
	UNIT_MILLISECOND
	UNIT_MICROSECOND
	UNIT_NANOSECOND
)

var x float64 = 365.242222 * 24 * float64(time.Hour)
var y float64 = math.Ceil(x / 12)
var timeUnitMap = map[int]int64{
	UNIT_HOUR:        int64(time.Hour),
	UNIT_MINUTE:      int64(time.Minute),
	UNIT_SECOND:      int64(time.Second),
	UNIT_MILLISECOND: int64(time.Millisecond),
	UNIT_MICROSECOND: int64(time.Microsecond),
	UNIT_NANOSECOND:  int64(time.Nanosecond),
	// naive
	UNIT_DAY: int64(time.Hour) * 24,
	//UNIT_MONTH: int64(y),
	//UNIT_YEAR:  int64(x),
	UNIT_MONTH: int64(365.25 * 24 * time.Hour / 12),
	UNIT_YEAR:  int64(365.25 * 24 * time.Hour),
}

var dateUnits = map[string]int{
	"y":       UNIT_YEAR,
	"year":    UNIT_YEAR,
	"quarter": UNIT_QUARTER,
	"month":   UNIT_MONTH,
	"M":       UNIT_MONTH,
	"week":    UNIT_WEEK,
	"w":       UNIT_WEEK,
	"day":     UNIT_DAY,
	"d":       UNIT_DAY,
}
var timeUnits = map[string]int{
	"hour":   UNIT_HOUR,
	"h":      UNIT_HOUR,
	"minute": UNIT_MINUTE,
	"m":      UNIT_MINUTE,
	"second": UNIT_SECOND,
	"s":      UNIT_SECOND,
	"ms":     UNIT_MILLISECOND,
	"micros": UNIT_MICROSECOND,
	"nanos":  UNIT_NANOSECOND,
}

func NewInterval(val string) Interval {
	return Interval{val: val, offsetDate: offsetDate{}}
}

type Interval struct {
	val        string
	date       time.Time
	offsetDate offsetDate
	offsetTime time.Duration
}

type offsetDate struct {
	Year    int64
	Month   int64
	Quarter int64
	Week    int64
	Day     int64
}

func (i *Interval) Parse() error {
	src := []byte(i.val)
	var s scanner.Scanner
	fset := token.NewFileSet()
	file := fset.AddFile("", fset.Base(), len(src))
	s.Init(file, src, nil /* no error handler */, 0)

	var (
		num    int64 = 1
		err    error
		offset int64
	)
	neg := false
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}

		switch tok {
		case token.SUB:
			neg = true
		case token.ADD:
			neg = false
		case token.INT:
			num, err = strconv.ParseInt(lit, 10, 32)
			if err != nil {
				return fmt.Errorf(`failed to parse interval: %s`, err)
			}
		case token.IDENT:
			v := num
			if neg {
				v *= -1
			}
			if dateUnit, ok := dateUnits[lit]; ok {
				switch dateUnit {
				case UNIT_YEAR:
					i.offsetDate.Year += v
				case UNIT_QUARTER:
					i.offsetDate.Quarter += v
				case UNIT_MONTH:
					i.offsetDate.Month += v
				case UNIT_WEEK:
					i.offsetDate.Week += v
				case UNIT_DAY:
					i.offsetDate.Day += v
				default:
					return fmt.Errorf(`unknown date unit: %s`, lit)
				}
			} else if timeUnit, ok := timeUnits[lit]; ok {
				if duration, ok := timeUnitMap[timeUnit]; ok {
					offset += v * duration
				} else {
					return fmt.Errorf(`duration is not set for time unit: %s`, lit)
				}
			} else {
				return fmt.Errorf(`not expected time unit: %s`, lit)
			}
			neg = false
			num = 1
		}
	}
	i.offsetTime = time.Duration(offset)
	return nil
}

func (i Interval) Apply(t time.Time) time.Time {
	// Align to year
	if i.offsetDate.Year == 1 {
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	}
	if i.offsetDate.Year > 1 { // use naive year offset
		v, _ := timeUnitMap[UNIT_YEAR]
		v_ := t.Truncate(time.Duration(i.offsetDate.Month * v))
		return time.Date(v_.Year(), 1, 1, 0, 0, 0, 0, v_.Location())
	}
	// Align to quarter
	if i.offsetDate.Quarter > 0 {
		mn := ((t.Month()-1)/3)*3 + 1
		return time.Date(t.Year(), mn, 1, 0, 0, 0, 0, t.Location())
	}
	// Align to week
	if i.offsetDate.Week > 0 {
		_, wn := t.ISOWeek()
		return time.Date(t.Year(), t.Month(), (wn-1)*7+1, 0, 0, 0, 0, t.Location())
	}
	// Align to month
	if i.offsetDate.Month == 1 {
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	}
	if i.offsetDate.Month > 1 { // use naive month offset
		v, _ := timeUnitMap[UNIT_MONTH]
		v_ := t.Truncate(time.Duration(i.offsetDate.Month * v))
		return time.Date(v_.Year(), v_.Month(), 1, 0, 0, 0, 0, v_.Location())
	}
	// Align to day
	if i.offsetDate.Day == 1 {
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	}
	if i.offsetDate.Day > 1 { // use naive day offset:w
		v, _ := timeUnitMap[UNIT_DAY]
		v_ := t.Truncate(time.Duration(i.offsetDate.Day * v))
		return time.Date(v_.Year(), v_.Month(), v_.Day(), 0, 0, 0, 0, v_.Location())
	}
	return t.Truncate(i.offsetTime)
}
