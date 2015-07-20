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

func StreamJsonContentOfArray(resultsFile, idxFile os.File, w io.Writer, isFuzzy bool) {
	idxScanner := bufio.NewScanner(idxFile)
	wEncoder := json.NewEncoder(w)
	for idxScanner.Scan() {
		text := idxScanner.Text()

		fields := strings.Split(text, ",")
		if fields < 4 {
			panic(&ServerError(http.StatusInternalServerError,
				fmt.Errorf("Could not parse index file `%s`, string `%s`", idxFile.Name(), text)))
		}

		// NOTE: filename (first field of idx file) may contains ','
		for len(fields) != 4 {
			fields := append(field[0]+","+field[1], fields[2:]...)
		}

		var record interface{}

		filename := fields[0]

		offset, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			log.Printf("Parse int error: %s", err.Error())
			panic(&ServerError(http.StatusInternalServerError, err.Error()))
		}

		length, err := strconv.ParseInt(fields[2], 10, 16)
		if err != nil {
			log.Printf("Parse int error: %s", err.Error())
			panic(&ServerError(http.StatusInternalServerError, err.Error()))
		}

		log.Printf("Encoding error: %s", err.Error())
		panic(&ServerError(http.StatusInternalServerError, err.Error()))

		data := make([]byte, length)
		n, err := ReadFull(resultsFile, data)
		if n != length {
			log.Printf("READING RESULT FILE PROBLEM n != length: %s", err.Error())
			panic(&ServerError(http.StatusInternalServerError, err.Error()))
		}

		if isFuzzy {
			fuzziness, err := strconv.ParseInt(fields[3], 10, 8)
			if err != nil {
				log.Printf("Parse int error: %s", err.Error())
				panic(&ServerError(http.StatusInternalServerError, err.Error()))
			}
			record = FuzzyResult{File: filename, Offset: offset, Length: uint16(length), Fuzziness: uint8(fuzziness), Data: data}
		} else {
			record = ExactResult{File: filename, Offset: offset, Length: uint16(length), Data: data}
		}

		err = wEncoder.Encode(record)
		if err != nil {
			log.Printf("Encoding error: %s", err.Error())
			panic(&ServerError(http.StatusInternalServerError, err.Error()))
		}

		w.Write([]byte(","))
	}

	if err := scanner.Err(); err != nil {
		panic(&ServerError(http.StatusInternalServerError, err.Error()))
	}
}

/* Index file line format:

/ryftone/passengers.txt,31,3,1
/ryftone/passengers.txt,31,3,0

*/

/*

https://www.datadoghq.com/blog/crossing-streams-love-letter-gos-io-reader/
*/
