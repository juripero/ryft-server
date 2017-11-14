package aggs

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/datetime"
)

// DateHist date_histogram engine
type DateHist struct {
	Field   utils.Field `json:"-" msgpack:"-"` // field path
	Missing interface{} `json:"-" msgpack:"-"` // missing value

	Interval datetime.Interval `json:"-" msgpack:"-"` // intervals like "month", "year" cannot be defined with time.Duration
	Offset   time.Duration     `json:"-" msgpack:"-"`
	Timezone *time.Location    `json:"-" msgpack:"-"`
	Format   string            `json:"-" msgpack:"-"`

	Buckets map[time.Time]*Bucket `json:"buckets,omitempty" msgpack:"buckets,omitempty"`

	// initial engine (prototype) that will be used for all buckets
	subAggs *Aggregations `json:"-" msgpack:"-"`
}

// data_histogram bucket
type Bucket struct {
	// number of records added to this bucket
	Count int64 `json:"count" msgpack:"count"`

	// sub-aggregations or nil
	SubAggs *Aggregations `json:"aggs,omitempty" msgpack:"aggs,omitempty"`
}

// add custom data to the bucker
func (b *Bucket) Add(data interface{}) error {
	if b.SubAggs != nil {
		// add already parsed data to bucket's engines
		for _, engine := range b.SubAggs.engines {
			if err := engine.Add(data); err != nil {
				return err
			}
		}
	}

	b.Count += 1
	return nil // OK
}

// merge the bucket (native)
func (b *Bucket) merge(other *Bucket) error {
	// merge sub-aggregations
	if b.SubAggs != nil {
		if err := b.SubAggs.Merge(other.SubAggs); err != nil {
			return err
		}
	}

	b.Count += other.Count
	return nil // OK
}

// merge the bucket (map)
func (b *Bucket) mergeMap(data_ interface{}) error {
	data, ok := data_.(map[string]interface{})
	if !ok {
		return fmt.Errorf("not a valid data")
	}

	// count is important
	count, err := utils.AsInt64(data["count"])
	if err != nil {
		return err
	}
	if count == 0 {
		return nil // nothing to merge
	}

	// merge sub-aggregations
	if b.SubAggs != nil {
		if err := b.SubAggs.Merge(data["aggs"]); err != nil {
			return err
		}
	}

	b.Count += count
	return nil // OK
}

// clone the engine
func (h *DateHist) clone() *DateHist {
	n := *h

	n.subAggs = h.subAggs.clone()
	n.Buckets = make(map[time.Time]*Bucket)
	for k, b := range h.Buckets {
		n.Buckets[k] = &Bucket{
			Count:   b.Count,
			SubAggs: b.SubAggs.clone(),
		}
	}

	return &n
}

// Name returns unique token for the current Engine
func (h *DateHist) Name() string {
	name := []string{
		fmt.Sprintf("field::%s", h.Field),
		fmt.Sprintf("interval::%s", h.Interval),
		fmt.Sprintf("format::%s", h.Format),
		fmt.Sprintf("timezone::%s", h.Timezone.String()),
	}

	// optional "missing" value
	if h.Missing != nil {
		name = append(name, fmt.Sprintf("missing::%s", h.Missing))
	}
	// optional "offset" value
	if h.Offset != 0 {
		name = append(name, fmt.Sprintf("offset::%s", h.Offset))
	}

	// names of all sub-aggregations
	if h.subAggs != nil {
		var subAggs []string
		for _, e := range h.subAggs.engines {
			subAggs = append(subAggs, e.Name())
		}
		name = append(name, fmt.Sprintf("sub-aggs<%s>",
			strings.Join(subAggs, "|")))
	}

	return fmt.Sprintf("datehist.%s", strings.Join(name, "/"))
}

// ToJson get object that can be serialized to JSON
func (h *DateHist) ToJson() interface{} {
	buckets := make(map[string]interface{})
	for k, b := range h.Buckets {
		bb := map[string]interface{}{
			"count": b.Count,
		}
		if b.SubAggs != nil {
			bb["aggs"] = b.SubAggs.ToJson(false)
		}
		buckets[k.Format(time.RFC3339)] = bb
	}

	return map[string]interface{}{
		"buckets": buckets,
	}
}

// get existing or create new bucket
func (h *DateHist) getBucket(key time.Time) *Bucket {
	b, ok := h.Buckets[key]
	if !ok {
		// create empty bucket
		b = &Bucket{
			SubAggs: h.subAggs.clone(),
		}

		if h.Buckets == nil {
			// create buckets container
			h.Buckets = make(map[time.Time]*Bucket)
		}

		h.Buckets[key] = b
	}

	return b
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

	val, err := parseDateTime(val_, h.Timezone, "")
	if err != nil {
		return fmt.Errorf("failed to parse datetime field: %s", err)
	}
	key := h.getBucketKey(val, h.Offset, h.Interval)

	// populate bucket
	bucket := h.getBucket(key.UTC())
	if err := bucket.Add(data); err != nil {
		return fmt.Errorf("sub-aggs failed: %s", err)
	}

	return nil // OK
}

func (h *DateHist) getBucketKey(val time.Time, offset time.Duration, interval datetime.Interval) time.Time {
	t := interval.Apply(val)
	t.Add(offset)
	return t
}

// Merge merge another aggregation engine
func (h *DateHist) Merge(data_ interface{}) error {
	switch data := data_.(type) {
	case *DateHist:
		return h.merge(data)
	case map[string]interface{}:
		return h.mergeMap(data)
	}

	return fmt.Errorf("no valid data")
}

// merge another intermediate aggregation (native)
func (h *DateHist) merge(other *DateHist) error {
	for k, b := range other.Buckets {
		bb := h.getBucket(k)
		if err := bb.merge(b); err != nil {
			return err
		}
	}

	return nil
}

// merge another intermediate aggregation (map)
func (h *DateHist) mergeMap(data map[string]interface{}) error {
	buckets, err := utils.AsStringMap(data["buckets"])
	if err != nil {
		return err
	}

	for kk, b := range buckets {
		k, err := time.Parse(time.RFC3339, kk)
		if err != nil {
			return err
		}

		bb := h.getBucket(k)
		if err := bb.mergeMap(b); err != nil {
			return err
		}
	}

	return nil // OK
}

// join another engine
func (h *DateHist) Join(other Engine) {
	// nothing to share
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
	sort.Sort(timeSlice(keys))

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

		keyAsString := datetime.FormatAsISO8601(f.engine.Format, k)

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
	field, err := getFieldOpt("field", opts, iNames)
	if err != nil {
		return nil, err
	}

	interval_, err := getStringOpt("interval", opts)
	if err != nil {
		return nil, err
	}
	if interval_ == "" {
		return nil, fmt.Errorf(`bad "interval": cannot be empty`)
	}

	interval := datetime.NewInterval(interval_)
	if err := interval.Parse(); err != nil {
		return nil, fmt.Errorf(`bad "interval": %s`, interval_)
	}

	timezone_, err := getStringOpt("timezone", opts)
	if err != nil {
		timezone_ = "UTC"
	}

	timezone, err := datetime.LoadTimezone(timezone_)
	if err != nil {
		return nil, fmt.Errorf(`bad "timezone": %s`, err)
	}

	format, err := getStringOpt("format", opts)
	if err != nil {
		// default key format
		format = "yyyy-MM-ddTHH:mm:ss.SSSZZ"
	}

	offset := time.Duration(0)
	if offset_, err := getStringOpt("offset", opts); err == nil {
		offset, err = parseSignedTimeUnitsInterval(offset_)
		if err != nil {
			return nil, fmt.Errorf(`bad "offset": %s`, err)
		}
	}

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
		Timezone: timezone,
		Format:   format,
		Offset:   offset,
		subAggs:  subAggs,
	}

	return &dateHistFunc{
		engine:      engine,
		keyed:       keyed,
		minDocCount: minDocCount,
	}, nil
}

// parse the date-time field
func parseDateTime(val interface{}, timezone *time.Location, formatHint string) (time.Time, error) {
	// get value as a string
	s, err := utils.AsString(val)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get datetime field: %s", err)
	}

	// convert string to timestamp
	t, err := datetime.ParseIn(s, timezone)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse datetime field: %s", err)
	}

	return t, nil // OK
}

// parseSignedTimeUnitsInterval process ElasticSearch time-units interval e.g.: "+6h" -> {6*time.Hour, error}; "-3m" -> {-3*time.Minute, error}
func parseSignedTimeUnitsInterval(v string) (time.Duration, error) {
	sign := int64(1)
	if strings.HasPrefix(v, "-") {
		sign = int64(-1)
	}
	interval, err := parseTimeUnitsInterval(strings.TrimLeft(v, "-+"))
	if err != nil {
		return 0, fmt.Errorf("failed to parse time units interval with sign: %s", err)
	}
	interval = time.Duration(int64(interval) * sign)
	return interval, nil
}

// parseTimeUnitsInterval parse ElasticSearch time-units interval
func parseTimeUnitsInterval(v string) (time.Duration, error) {
	var interval time.Duration
	compiledPattern, err := regexp.Compile(`^([\d]*)([\w]*)$`)
	if err != nil {
		return interval, fmt.Errorf(`failed to parse "interval" %s: %s`, v, err)
	}
	found := compiledPattern.FindAllStringSubmatch(v, -1)
	if len(found) == 0 || len(found[0]) < 3 {
		return interval, fmt.Errorf(`"interval" has wrong format %s`, v)
	}
	amount, err := utils.AsInt64(found[0][1])
	if err != nil {
		return interval, fmt.Errorf("failed to parse interval %s", v)
	}
	timeunit := found[0][2]
	switch timeunit {
	case "d":
		interval = time.Duration(int64(amount) * int64(24) * int64(time.Hour))
	case "h":
		interval = time.Duration(amount * int64(time.Hour))
	case "m":
		interval = time.Duration(amount * int64(time.Minute))
	case "s":
		interval = time.Duration(amount * int64(time.Second))
	case "ms":
		interval = time.Duration(amount * int64(time.Millisecond))
	case "micros":
		interval = time.Duration(amount * int64(time.Microsecond))
	case "nanos":
		interval = time.Duration(amount * int64(time.Nanosecond))
	default:
		return interval, fmt.Errorf("time-unit of interval set incorrectly %s", timeunit)
	}

	return interval, nil
}

// TimeSlice attaches the methods of sort.Interface to []time.Time, sorting in increasing order.
type timeSlice []time.Time

func (p timeSlice) Len() int           { return len(p) }
func (p timeSlice) Less(i, j int) bool { return p[i].Before(p[j]) }
func (p timeSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
