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

package datetime

import (
	"fmt"
	"go/scanner"
	"go/token"
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

var timeUnitMap = map[int]int64{
	UNIT_HOUR:        int64(time.Hour),
	UNIT_MINUTE:      int64(time.Minute),
	UNIT_SECOND:      int64(time.Second),
	UNIT_MILLISECOND: int64(time.Millisecond),
	UNIT_MICROSECOND: int64(time.Microsecond),
	UNIT_NANOSECOND:  int64(time.Nanosecond),
	UNIT_DAY:         int64(time.Hour) * 24,
	// naive
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

// Parse input date query
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
		// TODO: default should fail if met unexpected token
	}
	i.offsetTime = time.Duration(offset)
	return nil
}

// TimeUnitOffset return offset in time-units
func (i Interval) TimeUnitOffset() time.Duration {
	offset := time.Duration(i.offsetDate.Day * UNIT_DAY)
	offset += i.offsetTime
	return offset
}

// Truncate evalute date query and truncate date with the result offset
func (i Interval) Truncate(t time.Time) time.Time {
	// Align to year
	if i.offsetDate.Year == 1 {
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	}
	if i.offsetDate.Year > 1 { // use naive year offset
		v, _ := timeUnitMap[UNIT_YEAR]
		v_ := t.Truncate(time.Duration(i.offsetDate.Year * v))
		return time.Date(v_.Year(), 1, 1, 0, 0, 0, 0, v_.Location())
	}
	// Align to quarter
	if i.offsetDate.Quarter > 0 {
		mn := ((t.Month()-1)/3)*3 + 1
		return time.Date(t.Year(), mn, 1, 0, 0, 0, 0, t.Location())
	}
	// Align to week
	if i.offsetDate.Week > 0 {
		t := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		weekday := int(t.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		weekday = weekday - 1
		d := time.Duration(-weekday) * 24 * time.Hour
		t = t.Add(d)
		if i.offsetDate.Week == 1 {
			return t
		}
		return t.Truncate(time.Duration(UNIT_WEEK * i.offsetDate.Week))
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
	if i.offsetDate.Day > 1 { // use naive day offset
		v, _ := timeUnitMap[UNIT_DAY]
		v_ := t.Truncate(time.Duration(i.offsetDate.Day * v))
		return time.Date(v_.Year(), v_.Month(), v_.Day(), 0, 0, 0, 0, v_.Location())
	}
	return t.Truncate(i.offsetTime)
}
