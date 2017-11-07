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
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used
 *   to endorse or promote products derived from this software without specific prior written permission.
 *
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

package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// ParseDataRateMbps parses data rate in MB/s
// "inf" actually means that duration is zero (dataRate=length/duration)
// NOTE: need to sync all units with ryftprim!
func ParseDataRateMbps(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s) // case insensitive

	// trim suffix: KB, MB or GB
	scale := 1.0
	if t := strings.TrimSuffix(s, "kb/sec"); t != s {
		scale /= 1024
		s = t
	}
	if t := strings.TrimSuffix(s, "mb/sec"); t != s {
		// scale = 1.0
		s = t
	}
	if t := strings.TrimSuffix(s, "gb/sec"); t != s {
		scale *= 1024
		s = t
	}
	if t := strings.TrimSuffix(s, "tb/sec"); t != s {
		scale *= 1024 * 1024
		s = t
	}

	// parse data rate ("inf" is parsed as +Inf)
	s = strings.TrimSpace(s)
	r, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, err
	}

	// filter out any of +Int, -Inf, NaN
	if math.IsInf(r, 0) || math.IsNaN(r) {
		return 0.0, nil // report as zero!
	}

	return r * scale, nil // OK
}

// ParseDataSize parses total size in bytes.
// "inf" on "nan" mean zero
// NOTE: need to sync all units with ryftprim!
func ParseDataSize(x interface{}) (uint64, error) {
	// first try to parse as an integer
	tb, err := AsUint64(x)
	if err == nil {
		return tb, nil // OK
	}

	// then try to parse as a string
	s, err := AsString(x)
	if err != nil {
		return 0, err
	}
	s = strings.TrimSpace(s)
	s = strings.ToLower(s) // case insensitive

	// trim suffix: KB, MB or GB
	scale := uint64(1)
	if t := strings.TrimSuffix(s, "bytes"); t != s {
		// scale = 1
		s = t
	}
	if t := strings.TrimSuffix(s, "kb"); t != s {
		scale *= 1024
		s = t
	}
	if t := strings.TrimSuffix(s, "mb"); t != s {
		scale *= 1024 * 1024
		s = t
	}
	if t := strings.TrimSuffix(s, "gb"); t != s {
		scale *= 1024 * 1024
		scale *= 1024
		s = t
	}
	if t := strings.TrimSuffix(s, "tb"); t != s {
		scale *= 1024 * 1024
		scale *= 1024 * 1024
		s = t
	}

	s = strings.TrimSpace(s)
	if strings.ContainsAny(s, ".,einfa") {
		// value is float, parse as float64 ("inf" is parsed as +Inf)
		r, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, err
		}

		// filter out any of +Int, -Inf, NaN
		if math.IsInf(r, 0) || math.IsNaN(r) {
			return 0, nil // report as zero!
		}

		// TODO: check out of range
		return uint64(r * float64(scale)), nil // OK
	}

	// value is integer, parse as uint64!
	r, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}

	// TODO: check out of range
	return r * scale, nil // OK
}

// Formats defined in elasticsearch documentation (https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping-date-format.html)
const (
	BasicDateFormat                     string = "yyyyMMdd"
	BasicDateTimeFormat                 string = "yyyyMMdd'T'HHmmss.SSSZ"
	BasicDateTimeNoMillisFormat         string = "yyyyMMdd'T'HHmmssZ"
	BasicOriginalDateFormat             string = "yyyyDDD"
	BasicOriginalDateTimeFormat         string = "yyyyDDD'T'HHmmss.SSSZ"
	BasicOriginalDateTimeNoMillisFormat string = "yyyyDDD'T'HHmmssZ"
	BasicTimeFormat                     string = "HHmmss.SSSZ"
	BasicTimeNoMillisFormat             string = "HHmmssZ"
	BasicTTimeFormat                    string = "'T'HHmmss.SSSZ"
	BasicTTimeNoMillisFormat            string = "'T'HHmmssZ"
	BasicWeekDayFormat                  string = "xxxx'W'wwe"
	BasicWeekDayTimeFormat              string = "xxxx'W'wwe'T'HHmmss.SSSZ"
	BasicWeekDateTimeNoMillisFormat     string = "xxxx'W'wwe'T'HHmmssZ"
	DateFormat                          string = "yyyy-MM-dd"
	DateHourFormat                      string = "yyyy-MM-dd'T'HH"
	DateHourMinuteFormat                string = "yyyy-MM-dd'T'HH:mm"
	DateHourMinuteSecondFormat          string = "yyyy-MM-dd'T'HH:mm:ss"
	DateHourMinuteSecondFractionFormat  string = "yyyy-MM-dd'T'HH:mm:ss.SSS"
	DateHourMinuteSecondMillisFormat    string = "yyyy-MM-dd'T'HH:mm:ss.SSS"
	DateTimeFormat                      string = "yyyy-MM-dd'T'HH:mm:ss.SSSZZ"
	DateTimeNoMillisFormat              string = "yyyy-MM-dd'T'HH:mm:ssZZ"
	HourFormat                          string = "HH"
	HourMinuteFormat                    string = "HH:mm"
	HourMinuteSecondFormat              string = "HH:mm:ss"
	/*
		hour_minute_second_fraction or strict_hour_minute_second_fraction
		A formatter for a two digit hour of day, two digit minute of hour, two digit second of minute, and three digit fraction of second: HH:mm:ss.SSS.
		hour_minute_second_millis or strict_hour_minute_second_millis
		A formatter for a two digit hour of day, two digit minute of hour, two digit second of minute, and three digit fraction of second: HH:mm:ss.SSS.
		ordinal_date or strict_ordinal_date
		A formatter for a full ordinal date, using a four digit year and three digit dayOfYear: yyyy-DDD.
		ordinal_date_time or strict_ordinal_date_time
		A formatter for a full ordinal date and time, using a four digit year and three digit dayOfYear: yyyy-DDD'T'HH:mm:ss.SSSZZ.
		ordinal_date_time_no_millis or strict_ordinal_date_time_no_millis
		A formatter for a full ordinal date and time without millis, using a four digit year and three digit dayOfYear: yyyy-DDD'T'HH:mm:ssZZ.
		time or strict_time
		A formatter for a two digit hour of day, two digit minute of hour, two digit second of minute, three digit fraction of second, and time zone offset: HH:mm:ss.SSSZZ.
		time_no_millis or strict_time_no_millis
		A formatter for a two digit hour of day, two digit minute of hour, two digit second of minute, and time zone offset: HH:mm:ssZZ.
		t_time or strict_t_time
		A formatter for a two digit hour of day, two digit minute of hour, two digit second of minute, three digit fraction of second, and time zone offset prefixed by T: 'T'HH:mm:ss.SSSZZ.
		t_time_no_millis or strict_t_time_no_millis
		A formatter for a two digit hour of day, two digit minute of hour, two digit second of minute, and time zone offset prefixed by T: 'T'HH:mm:ssZZ.
		week_date or strict_week_date
		A formatter for a full date as four digit weekyear, two digit week of weekyear, and one digit day of week: xxxx-'W'ww-e.
		week_date_time or strict_week_date_time
		A formatter that combines a full weekyear date and time, separated by a T: xxxx-'W'ww-e'T'HH:mm:ss.SSSZZ.
		week_date_time_no_millis or strict_week_date_time_no_millis
		A formatter that combines a full weekyear date and time without millis, separated by a T: xxxx-'W'ww-e'T'HH:mm:ssZZ.
		weekyear or strict_weekyear
		A formatter for a four digit weekyear: xxxx.
		weekyear_week or strict_weekyear_week
		A formatter for a four digit weekyear and two digit week of weekyear: xxxx-'W'ww.
		weekyear_week_day or strict_weekyear_week_day
		A formatter for a four digit weekyear, two digit week of weekyear, and one digit day of week: xxxx-'W'ww-e.
		year or strict_year
		A formatter for a four digit year: yyyy.
		year_month or strict_year_month
		A formatter for a four digit year and two digit month of year: yyyy-MM.
		year_month_day or strict_year_month_day
		A formatter for a four digit year, two digit month of year, and two digit day of month: yyyy-MM-dd.
	*/
)

func ParseDate(v string) (time.Time, error) {
	dateformats := []string{
		BasicDateFormat,
		BasicDateTimeFormat,
		BasicDateTimeNoMillisFormat,
		BasicOriginalDateFormat,
		BasicOriginalDateTimeFormat,
		BasicOriginalDateTimeNoMillisFormat,
		BasicTimeFormat,
		BasicTimeNoMillisFormat,
		BasicTTimeFormat,
		BasicTTimeNoMillisFormat,
		BasicWeekDayFormat,
		BasicWeekDayTimeFormat,
		BasicWeekDateTimeNoMillisFormat,
		DateFormat,
		DateHourFormat,
		DateHourMinuteFormat,
		DateHourMinuteSecondFormat,
		DateHourMinuteSecondFractionFormat,
		DateHourMinuteSecondMillisFormat,
		DateTimeFormat,
		DateTimeNoMillisFormat,
		HourFormat,
		HourMinuteFormat,
		HourMinuteSecondFormat,
	}
	_ = dateformats
	for _, format := range dateformats {
		parsed, err := time.Parse(format, v)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("failed to get time from string %s", v)
}
