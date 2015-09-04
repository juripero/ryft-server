package universalxml

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func printResults(xml string) {
	out, err := DecodeBytes([]byte(xml))

	fmt.Printf("%s\n", xml)
	if err != nil {
		fmt.Printf("OUT ERR:%s", err.Error())
		return
	} else {

		b, err := json.Marshal(out)
		if err != nil {
			fmt.Printf("OUT MARSHAL ERR:%s", err.Error())
			return
		}

		var o bytes.Buffer
		fmt.Println("OUT:")
		json.Indent(&o, b, "", "  ")
		o.WriteTo(os.Stdout)
		fmt.Println("")
	}
}

func TestEasy(t *testing.T) {
	printResults(`<hello>hello world</hello>`)
}

func TestChilds(t *testing.T) {
	printResults(`<hello><a>aa</a><b>bb</b></hello>`)
}

func TestNamespaces(t *testing.T) {
	printResults(`<ns:a ns2:v='938293' s:r='4'><field1>value1</field1><field2>value2</field2></ns:a>`)
}
