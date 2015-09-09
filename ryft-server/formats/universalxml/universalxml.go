package universalxml

import (
	"bytes"
	"encoding/xml"
)

// https://jan.newmarch.name/go/xml/chapter-xml.html

func processToken(decoder *xml.Decoder) (obj map[string]interface{}, end bool, err error) {
	token, err := decoder.Token()

	if err != nil {
		return
	}

	switch t := token.(type) {
	case xml.CharData:
		// log.Printf("*** CharData: %s", string(xml.CharData(t)))
		obj = map[string]interface{}{
			"type": "chardata",
			"data": string(xml.CharData(t)),
		}
	case xml.Comment:
		obj = map[string]interface{}{
			"type": "comment",
			"data": string(xml.Comment(t)),
		}
	case xml.ProcInst:
		pi := xml.ProcInst(t)
		obj = map[string]interface{}{
			"type":   "procinst",
			"target": pi.Target,
			"data":   string(pi.Inst),
		}
	case xml.Directive:
		obj = map[string]interface{}{
			"type": "directive",
			"data": string(xml.Directive(t)),
		}
	case xml.EndElement:
		end = true
	case xml.StartElement:
		startElement := xml.StartElement(t)

		// log.Printf("*** Start: %s", startElement.Name.Local)

		var childs []interface{}
		for {
			child, childEnd, childErr := processToken(decoder)
			if childErr != nil {
				err = childErr
				return
			}

			if childEnd {
				break
			}

			childs = append(childs, child)
		}

		obj = map[string]interface{}{
			"type":      "element",
			"namespace": startElement.Name.Space,
			"name":      startElement.Name.Local,

			"attrs": func(attrs []xml.Attr) (array []map[string]string) {
				for _, a := range attrs {
					array = append(array, map[string]string{
						"namespace": a.Name.Space,
						"name":      a.Name.Local,
						"value":     a.Value,
					})
				}
				return
			}(startElement.Attr),
			"childs": childs,
		}
	}
	return
}

func DecodeBytes(raw []byte) (obj map[string]interface{}, err error) {
	buff := bytes.NewBuffer(raw)
	decoder := xml.NewDecoder(buff)
	obj, _, err = processToken(decoder)
	return
}
