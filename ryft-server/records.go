package main

import (
	"fmt"
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

func GetRecordsChan(idxFile *os.File, idxops chan fsnotify.Op, ch chan error) (records chan IdxRecord) {
	records = make(chan IdxRecord, 64)
	go func() {
		log.Printf("records: start records scanner")
	scan:
		for {
			for {
				var line string
				n, _ := fmt.Fscanln(idxFile, &line)

				if n == 0 {
					break // waiting for write event
				}

				r, err := NewIdxRecord(line)
				if err != nil {
					log.Printf("records: record parsing error '%s': %s", line, err.Error())
				}
				records <- r
				log.Printf("records: sent: %+v", r)
			}

		ops:
			for {
				select {
				case op := <-idxops:
					continue
				case err := <-ch:
					if err != nil {
						panic(err)
					} else {
						break scan
					}
				default:
					break ops
				}
			}
		}

		close(records)
	}()
	return
}
