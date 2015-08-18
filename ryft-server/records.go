package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/go-fsnotify/fsnotify"
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

func recordsScan(r io.Reader, records chan IdxRecord) error {
	log.Println("records-scan: start")

	var i uint64 = 0

	for {
		var line string
		n, _ := fmt.Fscanln(r, &line)
		if n == 0 {
			log.Println("records-scan: nothing for read -> complete ")
			break
		}

		r, err := NewIdxRecord(line)
		if err != nil {
			log.Printf("records-scan: record parsing error '%s': %s", line, err.Error())
			return err
		}

		log.Printf("records-scan: sending %s", line)
		records <- r
		log.Printf("records-scan: sent(%d): %+v", i, r)
		i++
	}
	log.Println("records-scan: end")
	return nil
}

func GetRecordsChan(idxFile *os.File, idxops chan fsnotify.Op, ch chan error) (records chan IdxRecord) {
	//records = make(chan IdxRecord, 64)
	records = make(chan IdxRecord, 4) // debugging purposes
	go func() {
	scan:
		for {
			if err := recordsScan(idxFile, records); err != nil {
				break scan
			}
		ops:
			for {
				select {
				case <-idxops:
					break ops
				case err := <-ch:
					if err != nil {
						panic(err)
					} else {
						recordsScan(idxFile, records)
						break scan
					}
				}
			}
		}

		close(records)
	}()
	return
}
