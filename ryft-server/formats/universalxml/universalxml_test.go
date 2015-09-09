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

func TestRealdR(t *testing.T) {
	const reald = `<rec><ID>10034183</ID><CaseNumber>HY223673</CaseNumber><Date>04/15/2015 11:59:00 PM</Date><Block>062XX S ST LAWRENCE AVE</Block><IUCR>0486</IUCR><PrimaryType>BATTERY</PrimaryType><Description>DOMESTIC BATTERY SIMPLE</Description><LocationDescription>STREET</LocationDescription><Arrest>false</Arrest><Domestic>true</Domestic><Beat>0313</Beat><District>003</District><Ward>20</Ward><CommunityArea>42</CommunityArea><FBICode>08B</FBICode><XCoordinate>1181263</XCoordinate><YCoordinate>1863965</YCoordinate><Year>2015</Year><UpdatedOn>04/22/2015 12:47:10 PM</UpdatedOn><Latitude>41.781961688</Latitude><Longitude>-87.610984705</Longitude><Location>"(41.781961688, -87.610984705)"</Location></rec>`
	printResults(reald)
}

func TestReald(t *testing.T) {
	const reald = `<rec><ID>10034183</ID><CaseNumber>HY223673</CaseNumber><Date>04/15/2015 11:59:00 PM</Date><Block>062XX S ST LAWRENCE AVE</Block><IUCR>0486</IUCR><PrimaryType>BATTERY</PrimaryType><Description>DOMESTIC BATTERY SIMPLE</Description><LocationDescription>STREET</LocationDescription><Arrest>false</Arrest><Domestic>true</Domestic><Beat>0313</Beat><District>003</District><Ward>20</Ward><CommunityArea>42</CommunityArea><FBICode>08B</FBICode><XCoordinate>1181263</XCoordinate><YCoordinate>1863965</YCoordinate><Year>2015</Year><UpdatedOn>04/22/2015 12:47:10 PM</UpdatedOn><Latitude>41.781961688</Latitude><Longitude>-87.610984705</Longitude><Location>"(41.781961688, -87.610984705)"</Location></rec`
	printResults(reald)
}
