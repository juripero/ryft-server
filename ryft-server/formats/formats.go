package formats

import (
	"log"

	"github.com/getryft/ryft-rest-api/ryft-server/formats/universalxml"
)

const XMLFormat = "xml"

var formats map[string]func(data []byte) (interface{}, error)

func Formats() map[string]func(data []byte) (interface{}, error) {
	return formats
}

func Available(name string) (hasParser bool) {
	if formats != nil {
		return
	}

	_, hasParser = formats[name]
	return
}

func Default() string {
	return XMLFormat
}

func init() {
	log.Printf("Formats init.")
	if formats != nil {
		return
	}
	formats = make(map[string]func(data []byte) (interface{}, error))
	formats[XMLFormat] = parseXML
}

func parseXML(data []byte) (interface{}, error) {
	return universalxml.DecodeBytes(data)
}
