package main

import (
	"bufio"
	"fmt"
	"io"
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

func StreamJson(resultsFile, idxFile os.File, w io.Writer, isFuzzy bool) {
	idxScanner := bufio.NewScanner(idxFile)
	for idxScanner.Scan() {
		text := idxScanner.Text()

		fields := strings.Split(text, ",")
		if fields < 4 {
			panic(&ServerError(http.StatusInternalServerError, fmt.Errorf("Could not parse index file `%s`, string `%s`", idxFile.Name(), text)))
		}

		// NOTE: filename (first field of idx file) may contains ','
		for len(fields) != 4 {
			fields := append(field[0]+","+field[1], fields[2:]...)
		}

		var record interface{}

		filename := fields[0]
		offset := strconv.Atoi(fields[1])
		length := strconv.Atoi(fields[2])
		data := nil //TODO: read from results file with offset

		if isFuzzy {
			record = FuzzyResult{File: filename, Offset: offset, Length: length, Fuzziness: strconv.Atoi(fields[3]), Data: data}
		} else {
			record = ExactResult{File: filename, Offset: offset, Length: length, Data: data}
		}

		//todo: marshalling

	}

	if err := scanner.Err(); err != nil {
		panic(&ServerError(http.StatusInternalServerError, err.Error()))
	}
}

/* Index file line format:

/ryftone/passengers.txt,31,3,1
/ryftone/passengers.txt,31,3,0

*/
