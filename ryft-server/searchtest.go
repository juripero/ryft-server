package main

import "net/http"

import "github.com/gin-gonic/gin"

const testresp interface{} = []interface{}{
	map[string]interface{}{
		"_index": map[string]interface{}{
			"file":      "/ryftone/passengers.txt",
			"offset":    211,
			"length":    25,
			"fuzziness": 0,
		},
		"field1": "value1",
		"field2": "value2",
		"field3": "value3",
		"field4": "value4",
	},
	map[string]interface{}{
		"_index": map[string]interface{}{
			"file":      "/ryftone/passengers.txt",
			"offset":    1211,
			"length":    25,
			"fuzziness": 0,
		},
		"field1": "value11",
		"field2": "value12",
		"field3": "value13",
		"field4": "value14",
	},
}

func searchtest(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, testresp)
}
