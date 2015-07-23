package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type IdxRecord struct {
	File      string `json:"file"`
	Offset    uint64 `json:"offset"`
	Length    uint16 `json:"length"`
	Fuzziness uint8  `json:"fuzziness"`
	Data      []byte `json:"data"`
}

func NewIdxRecord(line string) (r IdxRecord, err error) {
	fields := strings.Split(line, ",")
	if len(fields) < 4 {
		err = fmt.Errorf("Could not parse string `%s`", line)
		return
	}

	// NOTE: filename (first field of idx file) may contains ','
	for len(fields) != 4 {
		fields = append([]string{fields[0] + "," + fields[1]}, fields[2:]...)
	}

	r.File = fields[0]

	if r.Offset, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
		return
	}

	var length uint64
	if length, err = strconv.ParseUint(fields[2], 10, 16); err != nil {
		return
	}
	r.Length = uint16(length)

	var fuzziness uint64
	if fuzziness, err = strconv.ParseUint(fields[3], 10, 8); err != nil {
		return
	}
	r.Fuzziness = uint8(fuzziness)

	return
}

func StreamJson(resultsFile, idxFile *os.File, w io.Writer, completion chan error) {
	wEncoder := json.NewEncoder(w)
	idxRecords := make(chan IdxRecord, 64)
	go func() {
		for {
			select {
			case <-completion:
				recordsScan(idxFile, idxRecords)
				close(idxRecords)
				return
			default:
				recordsScan(idxFile, idxRecords)
			}
		}
	}()

	w.Write([]byte("["))
	defer w.Write([]byte("]"))

	var err error
	firstIteration := true
	for r := range idxRecords {
		if !firstIteration {
			w.Write([]byte(","))
		}

		err = wEncoder.Encode(r)
		if err != nil {
			log.Printf("Encoding error: %s", err.Error())
			panic(&ServerError{http.StatusInternalServerError, err.Error()})
		}

		firstIteration = false
	}
}

func recordsScan(r io.Reader, recordsChan chan IdxRecord) {
	for {
		var line string
		n, _ := fmt.Fscanln(r, &line)
		if n == 0 {
			break
		}

		r, err := NewIdxRecord(line)
		if err != nil {
			break
		}

		recordsChan <- r
	}
}

func linesScan(r io.Reader, linesChan chan string) {
	for {
		var line string
		n, _ := fmt.Fscanln(r, &line)
		if n == 0 {
			break
		}
		linesChan <- line
	}
}

func readDataBlock(r io.Reader, length uint16) (result []byte) {
	var total uint16 = 0
	for total < length {
		data := make([]byte, length-total)
		n, _ := r.Read(data)
		result = append(result, data...)
		total = total + uint16(n)
	}
	return
}

/* Index file line format:

/ryftone/passengers.txt,31,3,1
/ryftone/passengers.txt,31,3,0

*/

/*

https://www.datadoghq.com/blog/crossing-streams-love-letter-gos-io-reader/
*/
