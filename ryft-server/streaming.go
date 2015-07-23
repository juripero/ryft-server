package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// type ExactResult struct {
// 	File   string `json:"file"`
// 	Offset uint64 `json:"offset"`
// 	Length uint16 `json:"length"`
// 	Data   []byte `json:"data"`
// }

type IdxRecord struct {
	File      string `json:"file"`
	Offset    uint64 `json:"offset"`
	Length    uint16 `json:"length"`
	Fuzziness uint8  `json:"fuzziness"`
	Data      []byte `json:"data"`
}

func NewIdxRecord(line string) (r IdxRecord, err error) {
	fields := strings.Split(text, ",")
	if len(fields) < 4 {
		panic(&ServerError{http.StatusInternalServerError,
			fmt.Sprintf("Could not parse index file `%s`, string `%s`", idxFile.Name(), text)})
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
	r.Fuzziness = fuzziness

	return
}

// func StreamJsonContentOfArray(resultsFile, idxFile *os.File, w io.Writer, isFuzzy bool) {
// 	idxScanner := bufio.NewScanner(idxFile)
// 	wEncoder := json.NewEncoder(w)

// 	log.Println("+ IDXFILE:", idxFile)
// 	log.Println("+ IDXSCANNER:", idxScanner)
// 	log.Println("+ W:", w)
// 	log.Println("+ WENCODER:", wEncoder)

// 	w.Write([]byte("["))
// 	defer w.Write([]byte("]"))

// 	firstIteration := true
// 	log.Println("+ STARTING IDX ITERATIONS")
// 	for idxScanner.Scan() {
// 		log.Println("+ START NEW IDX ITERATION")
// 		if !firstIteration {
// 			w.Write([]byte(","))
// 		}

// 		var err error
// 		log.Println("+ BEGIN IDX RECORD")
// 		text := idxScanner.Text()
// 		log.Println("+ END IDX RECORD:", text)

// fields := strings.Split(text, ",")
// if len(fields) < 4 {
// 	panic(&ServerError{http.StatusInternalServerError,
// 		fmt.Sprintf("Could not parse index file `%s`, string `%s`", idxFile.Name(), text)})
// }

// // NOTE: filename (first field of idx file) may contains ','
// for len(fields) != 4 {
// 	fields = append([]string{fields[0] + "," + fields[1]}, fields[2:]...)
// }

// 		var record interface{}

// filename := fields[0]

// var offset uint64
// if offset, err = strconv.ParseUint(fields[1], 10, 64); err != nil {
// 	log.Printf("Parse int error: %s", err.Error())
// 	panic(&ServerError{http.StatusInternalServerError, err.Error()})
// }

// var length uint64
// if length, err = strconv.ParseUint(fields[2], 10, 16); err != nil {
// 	log.Printf("Parse int error: %s", err.Error())
// 	panic(&ServerError{http.StatusInternalServerError, err.Error()})
// }

// log.Printf("+ IDX PARSED: %s, %d, %d", filename, offset, length)

// 		data := make([]byte, length)
// 		n, err := io.ReadFull(resultsFile, data)

// 		// logging
// 		if err != nil {
// 			log.Printf("+ RES DATA:%s", string(data))
// 		}

// 		if n != int(length) {
// 			log.Printf("READING RESULT FILE PROBLEM n != length: %s", err.Error())
// 			panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 		}

// 		if isFuzzy {
// var fuzziness uint64
// if fuzziness, err = strconv.ParseUint(fields[3], 10, 8); err != nil {
// 	log.Printf("Parse int error: %s", err.Error())
// 	panic(&ServerError{http.StatusInternalServerError, err.Error()})
// }
// 			record = FuzzyResult{File: filename, Offset: offset, Length: uint16(length), Fuzziness: uint8(fuzziness), Data: data}
// 		} else {
// 			record = ExactResult{File: filename, Offset: offset, Length: uint16(length), Data: data}
// 		}

// 		err = wEncoder.Encode(record)
// 		if err != nil {
// 			log.Printf("Encoding error: %s", err.Error())
// 			panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 		}

// 		firstIteration = false

// 		log.Println("+ END NEW IDX ITERATION")
// 	}
// 	log.Println("+ END IDX ITERATIONS")

// 	if err := idxScanner.Err(); err != nil {
// 		panic(&ServerError{http.StatusInternalServerError, err.Error()})
// 	}
// }

func StreamJson(resultsFile, idxFile *os.File, w io.Writer, completion chan error) {
	//idxLines := make(chan string, 64)
	idxRecords := make(chan IdxRecord, 64)
	go func() {
		//idxScanner := bufio.NewScanner(idxFile)
		for {
			select {
			case <-completion:
				// linesScan(idxFile, idxLines)
				// close(idxLines)
				recordsScan(idxFile, idxRecords)
				return
			default:
				recordsScan(idxFile, idxRecords)
			}
		}
	}()

	// for line := range idxLines {
	// 	log.Printf("+ SCAN-LINE:%s", line)
	// }
	for r := range idxRecords {
		log.Printf("+ RECORD %+v", r)
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

func recordsScan(r io.Reader, recordsChan chan IdxRecord) {
	for {
		var line string
		n, _ := fmt.Fscanln(r, &line)
		if n == 0 {
			break
		}

		r, err := NewIdxRecord(line)
		if err {
			break
		}

		recordsChan <- r
	}
}

/* Index file line format:

/ryftone/passengers.txt,31,3,1
/ryftone/passengers.txt,31,3,0

*/

/*

https://www.datadoghq.com/blog/crossing-streams-love-letter-gos-io-reader/
*/
