package aggs

import (
	"fmt"
	"strings"
	"time"

	"github.com/getryft/ryft-server/search/utils"
)

type DateHist struct {
	Field    utils.Field `json:"-" msgpack:"-"`
	Interval string      `json:"-" msgpack:"-"`
	Timezone string      `json:"-" msgpack:"-"`
	Format   string      `json:"-" msgpack:"-"`

	flags int

	Engine  Engine                   `json:"engine" msgpack:"engine"` // initial engine that will be used for all buckets
	Buckets map[time.Duration]Engine `json:"buckets" msgpack:"buckets"`
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

// bucketName counts name of a bucket where an element should fall into
func (d DateHist) Key(data interface{}) (time.Duration, error) {
	return time.Duration(0), nil
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

		buckets = append(buckets, map[string]interface{}{
			"key_as_string": time.UTC.String(),
			"key":           name,
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
	format, _ := getStringOpt("format", opts)

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
			Buckets:  make(map[time.Duration]Engine),
		},
	}, nil
}
