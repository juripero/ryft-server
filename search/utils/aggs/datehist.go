package aggs

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/getryft/ryft-server/search/utils"
)

type DateHist struct {
	Field    utils.Field `json:"-" msgpack:"-"`
	Interval string      `json:"-" msgpack:"-"`
	Timezone string      `json:"-" msgpack:"-"`
	Format   string      `json:"-" msgpack:"-"`
	flags    int

	Engine  Engine               `json:"engine" msgpack:"engine"` // initial engine that will be used for all buckets
	Buckets map[time.Time]Engine `json:"buckets" msgpack:"buckets"`
}

// Name returns unique token for the current Engine
func (d *DateHist) Name() string {
	name := []string{
		fmt.Sprintf("field::%s", d.Field),
		fmt.Sprintf("interval::%s", d.Interval),
		fmt.Sprintf("timezone::%s", d.Timezone),
		fmt.Sprintf("format::%s", d.Format),
		fmt.Sprintf("engine::<<%s>>", d.Engine.Name()),
	}
	return fmt.Sprintf("datehist.%s", strings.Join(name, "/"))
}

// Key counts name of a bucket where an element should fall into
func (d DateHist) Key(data_ interface{}) (time.Time, error) {
	var key time.Time
	// find current `Field` and convert it into time.Time
	field := d.Field.String()
	data, err := utils.AsStringMap(data_)
	if err != nil {
		return key, fmt.Errorf("unable to create key: %s", err)
	}
	_, ok := data[field]
	if !ok {
		return key, fmt.Errorf("input data doesn't contain %s field", field)
	}
	v, err := utils.AsString(data[field])
	if err != nil {
		return key, fmt.Errorf("unable to create key: %s", err)
	}
	fieldDate, err := dateparse.ParseLocal(v)
	if err != nil {
		return key, fmt.Errorf("unable to create key: %s", err)
	}

	key, err = d.alignWithInterval(fieldDate)
	if err != nil {
		return key, fmt.Errorf("unable to create key: %s", err)
	}
	// find step (time.Time) and
	return key, nil
}

func (d DateHist) alignWithInterval(t time.Time) (time.Time, error) {
	var key time.Time
	interval := d.Field.String()
	// possible options: year, quarter, month, week, day, hour, minute, second
	if interval == "year" {
		key = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	} else if interval == "month" {
		key = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	} else if interval == "quarter" {
		month := t.Month()
		if month <= 3 {
			key = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
		} else if month <= 6 {
			key = time.Date(t.Year(), 4, 1, 0, 0, 0, 0, t.Location())
		} else if month <= 9 {
			key = time.Date(t.Year(), 7, 1, 0, 0, 0, 0, t.Location())
		} else if month <= 12 {
			key = time.Date(t.Year(), 10, 1, 0, 0, 0, 0, t.Location())
		} else {
			return key, fmt.Errorf("month number %s is out of [1:12] range", month)
		}
	} else if interval == "week" {
		_, week := t.ISOWeek()
		mondayAligned := (week-1)*7 + 1
		key = time.Date(t.Year(), t.Month(), mondayAligned, 0, 0, 0, 0, t.Location())
	} else if interval == "days" {
		key = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	} else if interval == "hour" {
		key = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	} else if interval == "minute" {
		key = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
	} else if interval == "second" {
		key = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, t.Location())
	} else {
		// if used time-units syntax (https://www.elastic.co/guide/en/elasticsearch/reference/current/common-options.html#time-units)
		interval = strings.ToLower(interval)
		interval = strings.Trim(interval, " +-")

		compiledPattern, err := regexp.Compile(`^([\d]*)([\w]*)$`)
		if err != nil {
			return key, fmt.Errorf("failed to parse interval %s", err)
		}
		found := compiledPattern.FindAllStringSubmatch(interval, -1)
		if len(found) == 0 || len(found[0]) < 3 {
			return key, fmt.Errorf("failed to parse interval %s", interval)
		}
		amount, err := utils.AsInt64(found[0][1])
		if err != nil {
			return key, fmt.Errorf("failed to parse interval %s", interval)
		}
		timeunit := found[0][2]
		switch timeunit {
		case "d":
			/*
				timeInterval := amount * int64(24) * int64(time.Hour)
				diff := t.Unix()
			*/

		case "h":
		case "s":
		case "ms":
		case "micros":
		case "nanos":
		default:
			return key, fmt.Errorf("time-unit of interval set incorrectly %s", timeunit)
		}

	}
	return key, nil
}

// ToJson get object that can be serialized to JSON
func (d *DateHist) ToJson() interface{} {
	return d
}

// Add add data to the aggregation
func (d *DateHist) Add(data interface{}) error {
	key, err := d.Key(data)
	if err != nil {
		return fmt.Errorf("record can't be assigned to any bucket: %s", err)
	}
	if _, ok := d.Buckets[key]; !ok {
		bucket := d.Engine
		d.Buckets[key] = bucket
	}
	// TODO: redefine stats struct in order to support date math (max, min, count(default))
	if err := d.Buckets[key].Add(map[string]interface{}{"Date": "1"}); err != nil {
		return fmt.Errorf("failed to add element %q into bucket %s: %s", data, key, err)
	}
	return nil
}

// Merge merge another aggregation engine
func (d *DateHist) Merge(data_ interface{}) error {
	switch data := data_.(type) {
	case *DateHist:
		return d.merge(data)
	case map[string]interface{}:
		return d.mergeMap(data)
	}

	return fmt.Errorf("no valid data")
}

func (d *DateHist) merge(other *DateHist) error {
	for k, engine := range other.Buckets {
		if _, ok := d.Buckets[k]; ok {
			d.Buckets[k].Merge(engine)
		} else {
			d.Buckets[k] = engine
		}
	}
	return nil
}

func (d *DateHist) mergeMap(data map[string]interface{}) error {
	return nil
}

// join another engine
func (d *DateHist) Join(other Engine) {
	if dd, ok := other.(*DateHist); ok {
		d.flags |= dd.flags
		// Field should be the same!
	}
}

type dateHistFunc struct {
	engine *DateHist
}

// ToJson
func (f *dateHistFunc) ToJson() interface{} {
	buckets := []interface{}{}
	for name, bucket := range f.engine.Buckets {
		data, ok := bucket.(*Stat)
		if !ok {
			continue
		}

		keyAsString := name.Format(f.engine.Format)
		buckets = append(buckets, map[string]interface{}{
			"key_as_string": keyAsString,
			"key":           name.UnixNano() / 1000000, // return in milliseconds since the epoch
			"doc_count":     data.Count,
		})
	}
	result := map[string][]interface{}{
		"buckets": buckets,
	}
	return result
}

// bind to another engine
func (f *dateHistFunc) bind(e Engine) {
	if d, ok := e.(*DateHist); ok {
		f.engine = d
	}
}

// clone function and engine
func (f *dateHistFunc) clone() (Function, Engine) {
	return nil, nil
}

func newDateHistFunc(opts map[string]interface{}, iNames []string) (*dateHistFunc, error) {
	field, err := getFieldOpt("field", opts, iNames)
	if err != nil {
		return nil, err
	}
	interval, err := getStringOpt("interval", opts)
	if err != nil {
		return nil, err
	}

	timezone, _ := getStringOpt("timezone", opts)
	format, err := getStringOpt("format", opts)
	if err != nil {
		format = time.RFC3339Nano
	}

	// TODO: extract sub-aggregations here
	countFunc, err := newCountFunc(opts, iNames)
	if err != nil {
		return nil, fmt.Errorf("unable to set Count sub-aggregation: %s", err)
	}

	return &dateHistFunc{
		engine: &DateHist{
			Field:    field,
			Interval: interval,
			Timezone: timezone,
			Format:   format,
			Engine:   countFunc.engine,
			Buckets:  make(map[time.Time]Engine),
		},
	}, nil
}
