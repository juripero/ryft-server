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

	// JSON data (JSON array format)
	assert.NoError(t, ioutil.WriteFile(filepath.Join(root, "data.jarr"),
		[]byte(`[
{"foo": {"bar": 100.0}} ,
{"foo": {"bar": "200"}} ,
{"foo": {"bar": 3e2}}
]`), 0644))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(root, "data.jarr.txt"),
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

	// do positive and negative tests
	check := func(n int, indexPath, dataPath, format string, opts string, expected string) {
		var params map[string]interface{}
		err := json.Unmarshal([]byte(opts), &params)
		assert.NoError(t, err)

		Aggs, err := aggs.MakeAggs(params, format, nil)
		if err != nil {
			assert.Contains(t, err.Error(), expected)
			return
		}

		err = ApplyAggregations(n, indexPath, dataPath, "\n", Aggs, true, nil)
		if err != nil {
			assert.Contains(t, err.Error(), expected)
		} else {
			outJson, err := json.Marshal(Aggs.ToJson(true))
			assert.NoError(t, err)

			assert.JSONEq(t, expected, string(outJson))
		}
	}

	all := true
	for n := 1; n < 16; n *= 2 {
		//if n := 2; true {

		// check JSON data
		if all {
			check(n, filepath.Join(root, "data.json.txt"), filepath.Join(root, "data.json"), "json",
				`{ "my": { "avg": { "field": "foo.bar" } } }`, `{"my": {"value": 200}}`)
			check(n, filepath.Join(root, "data.json.txt"), filepath.Join(root, "data.json"), "json",
				`{ "my": { "sum": { "field": "foo.bar" } } }`, `{"my": {"value": 600}}`)
			check(n, filepath.Join(root, "data.json.txt"), filepath.Join(root, "data.json"), "json",
				`{ "my": { "min": { "field": "foo.bar" } } }`, `{"my": {"value": 100}}`)
			check(n, filepath.Join(root, "data.json.txt"), filepath.Join(root, "data.json"), "json",
				`{ "my": { "max": { "field": "foo.bar" } } }`, `{"my": {"value": 300}}`)
			check(n, filepath.Join(root, "data.json.txt"), filepath.Join(root, "data.json"), "json",
				`{ "my": { "value_count": { "field": "foo.bar" } } }`, `{"my": {"value": 3}}`)
			check(n, filepath.Join(root, "data.json.txt"), filepath.Join(root, "data.json"), "json",
				`{ "my": { "stats": { "field": "foo.bar" } } }`, `{"my": {"avg": 200, "sum": 600, "min": 100, "max":300, "count": 3}}`)
		}

		// check JSON data (JSON array format)
		if all {
			check(n, filepath.Join(root, "data.jarr.txt"), filepath.Join(root, "data.jarr"), "json",
				`{ "my": { "avg": { "field": "foo.bar" } } }`, `{"my": {"value": 200}}`)
			check(n, filepath.Join(root, "data.jarr.txt"), filepath.Join(root, "data.jarr"), "json",
				`{ "my": { "sum": { "field": "foo.bar" } } }`, `{"my": {"value": 600}}`)
			check(n, filepath.Join(root, "data.jarr.txt"), filepath.Join(root, "data.jarr"), "json",
				`{ "my": { "min": { "field": "foo.bar" } } }`, `{"my": {"value": 100}}`)
			check(n, filepath.Join(root, "data.jarr.txt"), filepath.Join(root, "data.jarr"), "json",
				`{ "my": { "max": { "field": "foo.bar" } } }`, `{"my": {"value": 300}}`)
			check(n, filepath.Join(root, "data.jarr.txt"), filepath.Join(root, "data.jarr"), "json",
				`{ "my": { "value_count": { "field": "foo.bar" } } }`, `{"my": {"value": 3}}`)
			check(n, filepath.Join(root, "data.jarr.txt"), filepath.Join(root, "data.jarr"), "json",
				`{ "my": { "stats": { "field": "foo.bar" } } }`, `{"my": {"avg": 200, "sum": 600, "min": 100, "max":300, "count": 3}}`)
		}

		// check XML data
		if all {
			check(n, filepath.Join(root, "data.xml.txt"), filepath.Join(root, "data.xml"), "xml",
				`{ "my": { "avg": { "field": "foo.bar" } } }`, `{"my": {"value": 200}}`)
			check(n, filepath.Join(root, "data.xml.txt"), filepath.Join(root, "data.xml"), "xml",
				`{ "my": { "sum": { "field": "foo.bar" } } }`, `{"my": {"value": 600}}`)
			check(n, filepath.Join(root, "data.xml.txt"), filepath.Join(root, "data.xml"), "xml",
				`{ "my": { "min": { "field": "foo.bar" } } }`, `{"my": {"value": 100}}`)
			check(n, filepath.Join(root, "data.xml.txt"), filepath.Join(root, "data.xml"), "xml",
				`{ "my": { "max": { "field": "foo.bar" } } }`, `{"my": {"value": 300}}`)
			check(n, filepath.Join(root, "data.xml.txt"), filepath.Join(root, "data.xml"), "xml",
				`{ "my": { "value_count": { "field": "foo.bar" } } }`, `{"my": {"value": 3}}`)
			check(n, filepath.Join(root, "data.xml.txt"), filepath.Join(root, "data.xml"), "xml",
				`{ "my": { "stats": { "field": "foo.bar" } } }`, `{"my": {"avg": 200, "sum": 600, "min": 100, "max":300, "count": 3}}`)
		}

		// check UTF8 data
		if all {
			check(n, filepath.Join(root, "data.utf8.txt"), filepath.Join(root, "data.utf8"), "utf-8",
				`{ "my": { "avg": { "field": "." } } }`, `{"my": {"value": 200}}`)

			check(n, filepath.Join(root, "data.utf8.txt"), filepath.Join(root, "data.utf8"), "utf-8",
				`{ "my": { "sum": { "field": "." } } }`, `{"my": {"value": 600}}`)
			check(n, filepath.Join(root, "data.utf8.txt"), filepath.Join(root, "data.utf8"), "utf-8",
				`{ "my": { "min": { "field": "." } } }`, `{"my": {"value": 100}}`)
			check(n, filepath.Join(root, "data.utf8.txt"), filepath.Join(root, "data.utf8"), "utf-8",
				`{ "my": { "max": { "field": "." } } }`, `{"my": {"value": 300}}`)
			check(n, filepath.Join(root, "data.utf8.txt"), filepath.Join(root, "data.utf8"), "utf-8",
				`{ "my": { "value_count": { "field": "." } } }`, `{"my": {"value": 3}}`)
			check(n, filepath.Join(root, "data.utf8.txt"), filepath.Join(root, "data.utf8"), "utf-8",
				`{ "my": { "stats": { "field": "." } } }`, `{"my": {"avg": 200, "sum": 600, "min": 100, "max":300, "count": 3}}`)
		}
	}
}
