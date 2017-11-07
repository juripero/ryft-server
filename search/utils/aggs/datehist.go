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
	//Offset time.Duration
	//Timezone string        `json:"-" msgpack:"-"`
	//Format   string        `json:"-" msgpack:"-"`

	Buckets map[time.Time]*dateHistBucket `json:"buckets,omitempty" msgpack:"buckets,omitempty"`

	// initial engine (prototype) that will be used for all buckets
	subAggs *Aggregations `json:"-" msgpack:"-"`
}

// data_histogram bucket
type dateHistBucket struct {
	// number of records added to this bucket
	Count int64 `json:"count" msgpack:"count"`

	// sub-aggregations or nil
	SubAggs *Aggregations `json:"aggs,omitempty" msgpack:"aggs,omitempty"`
}

// clone the engine
func (h *DateHist) clone() *DateHist {
	n := *h

	n.subAggs = h.subAggs.clone()
	n.Buckets = make(map[time.Time]*dateHistBucket)
	for k, v := range h.Buckets {
		n.Buckets[k] = &dateHistBucket{
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
	val, err := parseDateTime(val_, "")
	if err != nil {
		return fmt.Errorf("failed to parse datetime field: %s", err)
	}

	// TODO: convert val to timezone and add custom offset!
	key := val.Truncate(h.Interval)
	key = key.UTC()

	// get bucket
	var bucket *dateHistBucket
	if b, ok := h.Buckets[key]; !ok {
		bucket = &dateHistBucket{
			Count:   0,
			SubAggs: h.subAggs.clone(),
		}

		// put it back
		if h.Buckets == nil {
			h.Buckets = make(map[time.Time]*dateHistBucket)
		}
		h.Buckets[key] = bucket
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

	// options
	keyed       bool
	minDocCount int64
}

// ToJson
func (f *dateHistFunc) ToJson() interface{} {
	keys := make([]time.Time, 0, len(f.engine.Buckets))
	for k, _ := range f.engine.Buckets {
		keys = append(keys, k)
	}
	sort.Sort(TimeSlice(keys))

	var ares []interface{}
	var mres map[string]interface{}
	if f.keyed {
		mres = make(map[string]interface{}, len(keys))
	} else {
		ares = make([]interface{}, 0, len(keys))
	}

	for _, k := range keys {
		bucket := f.engine.Buckets[k]
		if bucket.Count < f.minDocCount {
			continue
		}

		keyAsString := k.String() // .Format(f.engine.Format)
		b := map[string]interface{}{
			"key_as_string": keyAsString,
			"key":           k.UnixNano() / 1000000, // ns -> ms
			"doc_count":     bucket.Count,
		}
		if bucket.SubAggs != nil {
			subAggs := bucket.SubAggs.ToJson(true)
			for k, v := range subAggs {
				b[k] = v
			}
		}

		if f.keyed {
			mres[keyAsString] = b
		} else {
			ares = append(ares, b)
		}
	}

	var buckets interface{}
	if f.keyed {
		buckets = mres // JSON object
	} else {
		buckets = ares // JSON array
	}

	// TODO: extended_bounds etc...
	return map[string]interface{}{
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
	// field to get datetime from
	field, err := getFieldOpt("field", opts, iNames)
	if err != nil {
		return nil, err
	}

	// rounding interval
	interval_, err := getStringOpt("interval", opts)
	if err != nil {
		return nil, err
	}
	interval, err := parseInterval(interval_)
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

	// keyed
	var keyed bool
	if opt, ok := opts["keyed"]; ok {
		keyed, err = utils.AsBool(opt)
		if err != nil {
			return nil, fmt.Errorf(`bad "keyed" flag: %s`, err)
		}
	}

	// min doc count
	var minDocCount int64
	if opt, ok := opts["min_doc_count"]; ok {
		minDocCount, err = utils.AsInt64(opt)
		if err != nil {
			return nil, fmt.Errorf(`bad "min_doc_count" option: %s`, err)
		}
	}

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
		subAggs: subAggs,
	}

	return &dateHistFunc{
		engine:      engine,
		keyed:       keyed,
		minDocCount: minDocCount,
	}, nil
}

// parse the date-time field
func parseDateTime(val interface{}, formatHint string) (time.Time, error) {
	// get value as a string
	s, err := utils.AsString(val)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get datetime field: %s", err)
	}

	// convert string to timestamp
	t, err := dateparse.ParseLocal(s)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse datetime field: %s", err)
	}

	return t, nil // OK
}

// parse the interval option
func parseInterval(val string) (time.Duration, error) {
	// TODO: support for years, month etc...
	return time.ParseDuration(val)
}

// TimeSlice attaches the methods of sort.Interface to []time.Time, sorting in increasing order.
type TimeSlice []time.Time

func (p TimeSlice) Len() int           { return len(p) }
func (p TimeSlice) Less(i, j int) bool { return p[i].Before(p[j]) }
func (p TimeSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
