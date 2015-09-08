package formats

import (
	"log"

	"github.com/getryft/ryft-rest-api/ryft-server/formats/universalxml"
	"github.com/getryft/ryft-rest-api/ryft-server/records"
)

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
		log.Printf("Formats init.")
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
	log.Println("*** start xml convertor")
	obj, err := universalxml.DecodeBytes(r.Data)
	log.Println("**** xml convertor: bytes decoded")
	if err != nil {
		log.Println("**** xml convertor: bytes decoded with error: %s", err.Error())
		return nil, err
	}

	log.Println("**** xml convertor: start adding fields to obj=%+v", obj)
	addFields(obj, rawMap(r))
	log.Println("**** xml convertor: complete adding fields")

	log.Println("*** end xml convertor")
	return obj, nil
}

func addFields(m, from map[string]interface{}) {
	for k, v := range from {
		m[k] = v
	}
}

func rawMap(r records.IdxRecord) map[string]interface{} {
	var index = map[string]interface{}{
		"file":      r.File,
		"offset":    r.Offset,
		"length":    r.Length,
		"fuzziness": r.Fuzziness,
		"base64":    r.Data,
	}

	return map[string]interface{}{
		metaTag: index,
	}
}

func raw(r records.IdxRecord) (interface{}, error) {
	return rawMap(r), nil
}
