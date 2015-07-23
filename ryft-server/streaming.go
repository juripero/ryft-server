package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type ExactResult struct {
	File   string `json:"file"`
	Offset uint64 `json:"offset"`
	Length uint16 `json:"length"`
	Data   []byte `json:"data"`
}

type FuzzyResult struct {
	File      string `json:"file"`
	Offset    uint64 `json:"offset"`
	Length    uint16 `json:"length"`
	Fuzziness uint8  `json:"fuzziness"`
	Data      []byte `json:"data"`
}

func StreamJsonContentOfArray(resultsFile, idxFile *os.File, w io.Writer, isFuzzy bool) {
	idxScanner := bufio.NewScanner(idxFile)
	wEncoder := json.NewEncoder(w)

	w.Write([]byte("["))
	defer w.Write([]byte("]"))

	firstIteration := true
	for idxScanner.Scan() {
		if !firstIteration {
			w.Write([]byte(","))
		}

		var err error
		log.Println("+ BEGIN IDX RECORD")
		text := idxScanner.Text()
		log.Println("+ END IDX RECORD:", text)

		fields := strings.Split(text, ",")
		if len(fields) < 4 {
			panic(&ServerError{http.StatusInternalServerError,
				fmt.Sprintf("Could not parse index file `%s`, string `%s`", idxFile.Name(), text)})
		}

		// NOTE: filename (first field of idx file) may contains ','
		for len(fields) != 4 {
			fields = append([]string{fields[0] + "," + fields[1]}, fields[2:]...)
		}

		var record interface{}

		filename := fields[0]

		var offset uint64
		if offset, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
			log.Printf("Parse int error: %s", err.Error())
			panic(&ServerError{http.StatusInternalServerError, err.Error()})
		}

		var length uint64
		if length, err = strconv.ParseUint(fields[2], 10, 16); err != nil {
			log.Printf("Parse int error: %s", err.Error())
			panic(&ServerError{http.StatusInternalServerError, err.Error()})
		}

		log.Printf("+ IDX PARSED: %s, %d, %d", filename, offset, length)

		data := make([]byte, length)
		n, err := io.ReadFull(resultsFile, data)

		// logging
		if err != nil {
			log.Printf("+ RES DATA:%s", string(data))
		}

		if n != int(length) {
			log.Printf("READING RESULT FILE PROBLEM n != length: %s", err.Error())
			panic(&ServerError{http.StatusInternalServerError, err.Error()})
		}

		if isFuzzy {
			var fuzziness uint64
			if fuzziness, err = strconv.ParseUint(fields[3], 10, 8); err != nil {
				log.Printf("Parse int error: %s", err.Error())
				panic(&ServerError{http.StatusInternalServerError, err.Error()})
			}
			record = FuzzyResult{File: filename, Offset: offset, Length: uint16(length), Fuzziness: uint8(fuzziness), Data: data}
		} else {
			record = ExactResult{File: filename, Offset: offset, Length: uint16(length), Data: data}
		}

		err = wEncoder.Encode(record)
		if err != nil {
			log.Printf("Encoding error: %s", err.Error())
			panic(&ServerError{http.StatusInternalServerError, err.Error()})
		}

		firstIteration = false
	}

	if err := idxScanner.Err(); err != nil {
		panic(&ServerError{http.StatusInternalServerError, err.Error()})
	}
}

/* Index file line format:

/ryftone/passengers.txt,31,3,1
/ryftone/passengers.txt,31,3,0

*/

/*

https://www.datadoghq.com/blog/crossing-streams-love-letter-gos-io-reader/
*/
