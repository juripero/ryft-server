package aggs

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/getryft/ryft-server/search/utils"
)

// DateHist date_histogram engine
type DateHist struct {
	Field   utils.Field `json:"-" msgpack:"-"` // field path
	Missing interface{} `json:"-" msgpack:"-"` // missing value

	Interval time.Duration `json:"-" msgpack:"-"`
	Timezone string        `json:"-" msgpack:"-"`
	Format   string        `json:"-" msgpack:"-"`

	Buckets map[string]*dateHistBucket `json:"buckets,omitempty" msgpack:"buckets,omitempty"`

	// initial engine (prototype) that will be used for all buckets
	subAggs *Aggregations `json:"-" msgpack:"-"`
}

// data_histogram bucket
type dateHistBucket struct {
	Key time.Time `json:"key" msgpack:"key"`

	// number of records added to this bucket
	Count int64 `json:"count" msgpack:"count"`

	// sub-aggregations or nil
	SubAggs *Aggregations `json:"aggs,omitempty" msgpack:"aggs,omitempty"`
}

// clone the engine
func (h *DateHist) clone() *DateHist {
	n := *h

	n.subAggs = h.subAggs.clone()
	for k, v := range h.Buckets {
		n.Buckets[k] = &dateHistBucket{
			Key:     v.Key,
			Count:   v.Count,
			SubAggs: v.SubAggs.clone(),
		}
	}

	return &n
}

// Name returns unique token for the current Engine
func (h *DateHist) Name() string {
	name := []string{
		fmt.Sprintf("field::%s", h.Field),
		fmt.Sprintf("interval::%s", h.Interval),
		//fmt.Sprintf("timezone::%s", h.Timezone),
		//fmt.Sprintf("format::%s", h.Format),
	}
	//fmt.Sprintf("engine::%s", h.subAggs.Name()),

	return fmt.Sprintf("datehist.%s", strings.Join(name, "/"))
}

// ToJson get object that can be serialized to JSON
func (h *DateHist) ToJson() interface{} {
	return h
}

// Add add data to the aggregation
func (h *DateHist) Add(data interface{}) error {
	// extract field
	val_, err := h.Field.GetValue(data)
	if err != nil {
		if err == utils.ErrMissed {
			val_ = h.Missing // use provided value
		} else {
			return err
		}
	}
	if val_ == nil {
		return nil // do nothing if there is no value
	}

	// convert string to timestamp
	val, err := utils.AsString(val_)
	if err != nil {
		return fmt.Errorf("failed to get datetime field: %s", err)
	}
	ts, err := dateparse.ParseLocal(val)
	if err != nil {
		return fmt.Errorf("failed to parse datetime field: %s", err)
	}

	ts = ts.Truncate(h.Interval)
	key := ts.String()

	// get bucket
	var bucket *dateHistBucket
	if b, ok := h.Buckets[key]; !ok {
		bucket = &dateHistBucket{
			Key:     ts,
			Count:   0,
			SubAggs: h.subAggs.clone(),
		}
		h.Buckets[key] = bucket // put it back
	} else {
		bucket = b
	}

	if bucket.SubAggs != nil {
		// add already parsed data to bucket's engines
		for _, engine := range bucket.SubAggs.engines {
			if err := engine.Add(data); err != nil {
				return fmt.Errorf("sub-aggs failed: %s", err)
			}
		}
	}

	bucket.Count += 1
	return nil // OK
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

// merge another intermediate aggregation (native)
func (d *DateHist) merge(other *DateHist) error {
	/*for k, engine := range other.Buckets {
		if _, ok := d.Buckets[k]; ok {
			d.Buckets[k].Merge(engine)
		} else {
			d.Buckets[k] = engine
		}
	}
	return nil */
	return fmt.Errorf("merge is not implemented YET")
}

// merge another intermediate aggregation (map)
func (d *DateHist) mergeMap(data map[string]interface{}) error {
	/*return nil*/
	return fmt.Errorf("merge is not implemented YET")
}

// join another engine
func (d *DateHist) Join(other Engine) {
	panic(fmt.Errorf("join is not implemented YET"))
}

// "date_histogram" aggregation function
type dateHistFunc struct {
	engine *DateHist
}

// ToJson
func (f *dateHistFunc) ToJson() interface{} {
	keys := make([]string, 0, len(f.engine.Buckets))
	for k, _ := range f.engine.Buckets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buckets []interface{}
	for _, k := range keys {
		bucket := f.engine.Buckets[k]
		keyAsString := "" //bucket.Key.String() // .Format(f.engine.Format)
		buckets = append(buckets,
			map[string]interface{}{
				"key_as_string": keyAsString,
				"key":           bucket.Key.UnixNano() / 1000000, // ns -> ms
				"doc_count":     bucket.Count,
			})
	}

	// TODO: keyed, min_doc_count, extended_bounds etc...
	return map[string][]interface{}{
		"buckets": buckets,
	}
}

// bind to another engine
func (f *dateHistFunc) bind(e Engine) {
	if d, ok := e.(*DateHist); ok {
		f.engine = d
	}
}

// clone function and engine
func (f *dateHistFunc) clone() (Function, Engine) {
	n := &dateHistFunc{}
	n.engine = f.engine.clone() // copy engine
	return n, n.engine
}

// make new "date_histrogram" aggregation
func newDateHistFunc(opts map[string]interface{}, iNames []string) (*dateHistFunc, error) {
	field, err := getFieldOpt("field", opts, iNames)
	if err != nil {
		return nil, fmt.Errorf(`bad "field": %s`, err)
	}
	interval_, err := getStringOpt("interval", opts)
	if err != nil {
		return nil, fmt.Errorf(`bad "interval": %s`, err)
	}
	interval, err := time.ParseDuration(interval_)
	if err != nil {
		return nil, fmt.Errorf(`bad "interval": %s`, err)
	}

	/*
		timezone, _ := getStringOpt("timezone", opts)
		format, err := getStringOpt("format", opts)
		if err != nil {
			format = time.RFC3339Nano
		}
	*/

	// parse sub-aggregations
	var subAggs *Aggregations
	if aggs_, ok := opts[AGGS_NAME]; ok {
		aggsOpts, err := utils.AsStringMap(aggs_)
		if err != nil {
			return nil, fmt.Errorf("failed to get sub-aggregation: %s", err)
		}

		subAggs, err = makeAggs(aggsOpts, "-", nil)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sub-aggregation: %s", err)
		}
	}

	// main engine
	engine := &DateHist{
		Field:    field,
		Missing:  opts["missing"],
		Interval: interval,
		//Timezone: timezone,
		//Format:   format,
		Buckets: make(map[string]*dateHistBucket),
		subAggs: subAggs,
	}

	return &dateHistFunc{
		engine: engine,
	}, nil
}
