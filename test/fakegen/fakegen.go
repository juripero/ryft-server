// generate fake data
/*
./fakegen --count=500 --pattern '${rand(64)} hello world ${rand(64)}' > /ryftone/fake.txt

curl -s 'http://localhost:8765/search?query=(RAW_TEXT%20CONTAINS%20"hello%20world")&files=fake10.txt&stats=true&mode=fhs&fuzziness=1' | jq -c '.results | sort_by(._index.offset) | .[]._index'
curl -s 'http://localhost:8765/search?query=(RAW_TEXT%20CONTAINS%20"hello%20world")AND(RAW_TEXT%20CONTAINS%20FHS("hello%20world",DIST=0))&files=fake10.txt&stats=true&mode=fhs&fuzziness=1' | jq -c '.results | sort_by(._index.offset) | .[]._index'
*/
package main

import (
	"bytes"
	"fmt"
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
	re := regexp.MustCompile(`\$\{rand\((.+?)\)\}`)

	for i := 0; i < count; i++ {
		parts := []string{}
		pos := 0

		match := re.FindAllStringSubmatchIndex(pattern, -1)
		for m := 0; m < len(match); m++ {
			beg := match[m][0]
			end := match[m][1]
			args := pattern[match[m][2]:match[m][3]]

			if pos < beg {
				parts = append(parts, pattern[pos:beg])
			}
			parts = append(parts, randstr(args))
			pos = end
		}

		parts = append(parts, pattern[pos:], "\n")
		os.Stdout.WriteString(strings.Join(parts, ""))
	}
}

// generate random string
func randstr(inArgs string) string {
	pattern := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	min := 0
	max := 0

	args := strings.Split(inArgs, ",")
	for i, s := range args {
		args[i] = strings.TrimSpace(s)
	}
	// fmt.Println(args)
	switch len(args) {
	case 1:
		if n, err := strconv.ParseInt(args[0], 10, 32); err != nil {
			panic(err)
		} else {
			min = int(n)
			max = int(n)
		}

	case 2:
		if n, err := strconv.ParseInt(args[0], 10, 32); err != nil {
			panic(err)
		} else {
			min = int(n)
		}
		if n, err := strconv.ParseInt(args[1], 10, 32); err != nil {
			pattern = args[1]
		} else {
			max = int(n)
		}

	case 3:
		if n, err := strconv.ParseInt(args[0], 10, 32); err != nil {
			panic(err)
		} else {
			min = int(n)
		}
		if n, err := strconv.ParseInt(args[1], 10, 32); err != nil {
			panic(err)
		} else {
			max = int(n)
		}
		pattern = args[2]

	default:
		panic(fmt.Errorf("invalid number of arguments: %s", inArgs))
	}

	if max < min {
		min, max = max, min
	}
	if len(pattern) == 0 {
		panic(fmt.Errorf("no valid pattern provided"))
	}

	buf := new(bytes.Buffer)
	n := min
	if max-min > 0 {
		n += rand.Intn(max - min)
	}
	for i := 0; i < n; i++ {
		k := rand.Intn(len(pattern))
		buf.WriteByte(pattern[k])
	}
	return buf.String()
}