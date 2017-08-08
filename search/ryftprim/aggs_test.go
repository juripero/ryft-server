package ryftprim

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search/utils/aggs"
	"github.com/stretchr/testify/assert"
)

// test aggregations
func TestApplyAggregations(t *testing.T) {
	SetLogLevelString(testLogLevel)

	root := fmt.Sprintf("/tmp/ryft-%x", time.Now().UnixNano())
	assert.NoError(t, os.MkdirAll(root, 0755))
	defer os.RemoveAll(root)

	// JSON data
	assert.NoError(t, ioutil.WriteFile(filepath.Join(root, "data.json"),
		[]byte(`{"foo": {"bar": 100.0}}
{"foo": {"bar": "200"}}
{"foo": {"bar": 3e2}}
`), 0644))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(root, "data.json.txt"),
		[]byte(`1.json,1,23,0
2.json,2,23,0
3.json,3,21,0`), 0644))

	// XML data
	assert.NoError(t, ioutil.WriteFile(filepath.Join(root, "data.xml"),
		[]byte(`<rec><foo><bar>100.0</bar></foo></rec>
<rec><foo><bar> 200 </bar></foo></rec>
<rec><foo><bar>3e2</bar></foo></rec>
`), 0644))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(root, "data.xml.txt"),
		[]byte(`1.xml,1,38,0
2.xml,2,38,0
3.xml,3,36,0`), 0644))

	// UTF-8 numbers
	assert.NoError(t, ioutil.WriteFile(filepath.Join(root, "data.utf8"),
		[]byte(`100.0
200
3e2
`), 0644))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(root, "data.utf8.txt"),
		[]byte(`1.txt,1,5,0
2.txt,2,3,0
3.txt,3,3,0`), 0644))

	assert.NoError(t, ioutil.WriteFile(filepath.Join(root, "data.geo.xml"),
		[]byte(`
		<rec><Latitude>10.000000000</Latitude><Longitude>10.000000000</Longitude><Location>"(10.000000000, 10.000000000)"</Location></rec>
		<rec><Latitude>30.000000000</Latitude><Longitude>-20.000000000</Longitude><Location>"(30.000000000, -20.000000000)"</Location></rec>
		<rec><Latitude>40.000000000</Latitude><Longitude>-30.000000000</Longitude><Location>"(40.000000000, -30.000000000)"</Location></rec>
`), 0644))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(root, "data.geo.xml.txt"),
		[]byte(`1.txt,1,30,0
2.txt,2,30,0,
3.txt,3,30,0`), 0644))

	// do positive and negative tests
	check := func(indexPath, dataPath, format string, aggregations map[string]map[string]map[string]interface{}, expected string) {
		Aggs, err := aggs.MakeAggs(aggregations)
		if err != nil {
			assert.Contains(t, err.Error(), expected)
			return
		}

		err = ApplyAggregations(indexPath, dataPath, "\n", format, Aggs)
		if err != nil {
			assert.Contains(t, err.Error(), expected)
		} else {
			outJson, err := json.Marshal(Aggs.ToJson(true))
			assert.NoError(t, err)

			assert.JSONEq(t, expected, string(outJson))
		}
	}

	if true {
		check(filepath.Join(root, "data.geo.xml.txt"), filepath.Join(root, "data.xml"), "xml",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"geo_bounds": map[string]interface{}{
						"field": "rec",
					},
				},
			}, `{"my": {"geo_bounds": {"top_left": {"lat": 0.0, "lon": 0.0}, "bottom_right": {"lat": 0.0, "lon": 0.0}}}}`)

	}

	// check JSON data
	if true {
		check(filepath.Join(root, "data.json.txt"), filepath.Join(root, "data.json"), "json",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"avg": map[string]interface{}{
						"field": "foo.bar",
					},
				},
			}, `{"my": {"value": 200}}`)
		check(filepath.Join(root, "data.json.txt"), filepath.Join(root, "data.json"), "json",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"sum": map[string]interface{}{
						"field": "foo.bar",
					},
				},
			}, `{"my": {"value": 600}}`)
		check(filepath.Join(root, "data.json.txt"), filepath.Join(root, "data.json"), "json",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"min": map[string]interface{}{
						"field": "foo.bar",
					},
				},
			}, `{"my": {"value": 100}}`)
		check(filepath.Join(root, "data.json.txt"), filepath.Join(root, "data.json"), "json",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"max": map[string]interface{}{
						"field": "foo.bar",
					},
				},
			}, `{"my": {"value": 300}}`)
		check(filepath.Join(root, "data.json.txt"), filepath.Join(root, "data.json"), "json",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"value_count": map[string]interface{}{
						"field": "foo.bar",
					},
				},
			}, `{"my": {"value": 3}}`)
		check(filepath.Join(root, "data.json.txt"), filepath.Join(root, "data.json"), "json",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"stat": map[string]interface{}{
						"field": "foo.bar",
					},
				},
			}, `{"my": {"avg": 200, "sum": 600, "min": 100, "max":300, "count": 3}}`)
	}

	// check XML data
	if true {
		check(filepath.Join(root, "data.xml.txt"), filepath.Join(root, "data.xml"), "xml",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"avg": map[string]interface{}{
						"field": "foo.bar",
					},
				},
			}, `{"my": {"value": 200}}`)
		check(filepath.Join(root, "data.xml.txt"), filepath.Join(root, "data.xml"), "xml",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"sum": map[string]interface{}{
						"field": "foo.bar",
					},
				},
			}, `{"my": {"value": 600}}`)
		check(filepath.Join(root, "data.xml.txt"), filepath.Join(root, "data.xml"), "xml",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"min": map[string]interface{}{
						"field": "foo.bar",
					},
				},
			}, `{"my": {"value": 100}}`)
		check(filepath.Join(root, "data.xml.txt"), filepath.Join(root, "data.xml"), "xml",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"max": map[string]interface{}{
						"field": "foo.bar",
					},
				},
			}, `{"my": {"value": 300}}`)
		check(filepath.Join(root, "data.xml.txt"), filepath.Join(root, "data.xml"), "xml",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"value_count": map[string]interface{}{
						"field": "foo.bar",
					},
				},
			}, `{"my": {"value": 3}}`)
		check(filepath.Join(root, "data.xml.txt"), filepath.Join(root, "data.xml"), "xml",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"stat": map[string]interface{}{
						"field": "foo.bar",
					},
				},
			}, `{"my": {"avg": 200, "sum": 600, "min": 100, "max":300, "count": 3}}`)
	}

	// check UTF8 data
	if true {
		check(filepath.Join(root, "data.utf8.txt"), filepath.Join(root, "data.utf8"), "utf-8",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"avg": map[string]interface{}{
						"field": ".",
					},
				},
			}, `{"my": {"value": 200}}`)

		check(filepath.Join(root, "data.utf8.txt"), filepath.Join(root, "data.utf8"), "utf-8",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"sum": map[string]interface{}{
						"field": ".",
					},
				},
			}, `{"my": {"value": 600}}`)
		check(filepath.Join(root, "data.utf8.txt"), filepath.Join(root, "data.utf8"), "utf-8",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"min": map[string]interface{}{
						"field": ".",
					},
				},
			}, `{"my": {"value": 100}}`)
		check(filepath.Join(root, "data.utf8.txt"), filepath.Join(root, "data.utf8"), "utf-8",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"max": map[string]interface{}{
						"field": ".",
					},
				},
			}, `{"my": {"value": 300}}`)
		check(filepath.Join(root, "data.utf8.txt"), filepath.Join(root, "data.utf8"), "utf-8",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"value_count": map[string]interface{}{
						"field": ".",
					},
				},
			}, `{"my": {"value": 3}}`)
		check(filepath.Join(root, "data.utf8.txt"), filepath.Join(root, "data.utf8"), "utf-8",
			map[string]map[string]map[string]interface{}{
				"my": map[string]map[string]interface{}{
					"stat": map[string]interface{}{
						"field": ".",
					},
				},
			}, `{"my": {"avg": 200, "sum": 600, "min": 100, "max":300, "count": 3}}`)
	}
}
