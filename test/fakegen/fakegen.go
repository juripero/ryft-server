// generate fake data
/*
./fakegen --count=500 --pattern '${rand(64)} hello world ${rand(64)}' > /ryftone/fake.txt

curl -s 'http://localhost:8765/search?query=(RAW_TEXT%20CONTAINS%20"hello%20world")&files=fake10.txt&stats=true&mode=fhs&fuzziness=1' | jq -c '.results | sort_by(._index.offset) | .[]._index'
curl -s 'http://localhost:8765/search?query=(RAW_TEXT%20CONTAINS%20"hello%20world")AND(RAW_TEXT%20CONTAINS%20FHS("hello%20world",DIST=0))&files=fake10.txt&stats=true&mode=fhs&fuzziness=1' | jq -c '.results | sort_by(._index.offset) | .[]._index'
*/
package main

import (
	"bytes"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var pattern string
	var count int

	kingpin.Flag("pattern", "pattern for record").Short('p').StringVar(&pattern)
	kingpin.Flag("count", "number of records").Short('n').Default("10").IntVar(&count)
	kingpin.Parse()

	rand.Seed(time.Now().UnixNano())
	re := regexp.MustCompile(`\$\{rand\((\d+)\)\}`)

	for i := 0; i < count; i++ {
		parts := []string{}
		pos := 0

		match := re.FindAllStringSubmatchIndex(pattern, -1)
		for m := 0; m < len(match); m++ {
			beg := match[m][0]
			end := match[m][1]
			n := pattern[match[m][2]:match[m][3]]

			if pos < beg {
				parts = append(parts, pattern[pos:beg])
			}
			parts = append(parts, randstr(n))
			pos = end
		}

		parts = append(parts, pattern[pos:], "\n")
		os.Stdout.WriteString(strings.Join(parts, ""))
	}
}

// generate random string
func randstr(s string) string {
	pattern := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	n, _ := strconv.ParseInt(s, 10, 32)
	buf := new(bytes.Buffer)
	for i := int64(0); i < n; i++ {
		k := rand.Intn(len(pattern))
		buf.WriteByte(pattern[k])
	}
	return buf.String()
}
