package formats

import (
	"fmt"

	"github.com/getryft/ryft-rest-api/ryft-server/records"
)
import "github.com/clbanning/x2j"

const (
	XMLFormat = "xml"
	RAWFormat = "raw"
)

const (
	metaTag = "_index"
)

var formats map[string]func(r records.IdxRecord) (interface{}, error)

func Formats() map[string]func(r records.IdxRecord) (interface{}, error) {
	if formats == nil {
		formats = make(map[string]func(r records.IdxRecord) (interface{}, error))
		formats[XMLFormat] = xml
		formats[RAWFormat] = raw
	}

	return formats
}

func Available(name string) (hasParser bool) {
	_, hasParser = Formats()[name]
	return
}

func Default() string {
	return RAWFormat
}

func xml(r records.IdxRecord) (interface{}, error) {
	obj, err := x2j.ByteDocToMap(r.Data, false)
	if err != nil {
		return nil, err
	}
	data, ok := obj["rec"]
	if ok {
		addFields(data.(map[string]interface{}), rawMap(r, true))
		return data, nil
	} else {
		return nil, fmt.Errorf("Could not parse xml")
	}

}

func addFields(m, from map[string]interface{}) {
	for k, v := range from {
		m[k] = v
	}
}

func rawMap(r records.IdxRecord, isXml bool) map[string]interface{} {
	var index = map[string]interface{}{
		"file":      r.File,
		"offset":    r.Offset,
		"length":    r.Length,
		"fuzziness": r.Fuzziness,
	}
	if isXml {
		return map[string]interface{}{
			metaTag: index,
		}
	} else {
		return map[string]interface{}{
			metaTag:  index,
			"base64": r.Data,
		}
	}
}

func raw(r records.IdxRecord) (interface{}, error) {
	return rawMap(r, false), nil
}
