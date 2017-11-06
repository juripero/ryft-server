package aggs

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testPopulate(t *testing.T, engine Engine) {
	assert.NoError(t, engine.Add(map[string]interface{}{"Data": "1", "Date": "04/15/2015 11:59:00 PM", "UpdatedOn": "04/22/2015 12:47:10 PM", "Year": "2015"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Data": "2", "Date": "04/15/2015 11:55:00 PM", "UpdatedOn": "04/22/2015 12:47:10 PM", "Year": "2015"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Data": "3", "Date": "05/15/2015 11:55:00 PM", "UpdatedOn": "06/22/2015 12:47:10 PM", "Year": "2015"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Data": "4", "Date": "05/20/2015 11:55:00 PM", "UpdatedOn": "06/22/2015 12:47:10 PM", "Year": "2015"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Data": "5", "Date": "11/01/2016 11:55:00 PM", "UpdatedOn": "01/01/2017 12:47:10 PM", "Year": "2016"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Data": "6", "Date": "01/02/2017 11:55:00 PM", "UpdatedOn": "01/02/2017 12:47:10 PM", "Year": "2017"}))
	assert.NoError(t, engine.Add(map[string]interface{}{"Data": "7", "Date": "02/02/2017 07:00:00 AM", "UpdatedOn": "02/02/2017 07:00:49 PM", "Year": "2017"}))
}

func TestDateHistogramEngine(t *testing.T) {
	opts := map[string]interface{}{
		"field":    "Date",
		"interval": "1d",
	}
	fn, err := newDateHistFunc(opts, nil)
	assert.NoError(t, err)
	testPopulate(t, fn.engine)
	res := fn.ToJson()
	log.Printf("result %s", res)
}
