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

func recordsScan(r io.Reader, records chan IdxRecord, dropper chan struct{}) error {

	var i uint64 = 0

	for {
		var line string

		select {
		case <-dropper:
			log.Println("records: read - external termination")
			return fmt.Errorf("External termination")
		default:
		}

		n, _ := fmt.Fscanln(r, &line)
		if n == 0 {
			break
		}

		r, err := NewIdxRecord(line)
		if err != nil {
			log.Printf("records: error '%s': %s", line, err.Error())
			return err
		}

		records <- r
		i++
	}
	return nil
}

func GetRecordsChan(idxFile *os.File, idxops chan fsnotify.Op, ch chan error, dropper chan struct{}) (records chan IdxRecord) {
	records = make(chan IdxRecord, 4)
	go func() {
	scan:
		for {
			if err := recordsScan(idxFile, records, dropper); err != nil {
				break scan
			}
		ops:
			for {
				select {
				case <-dropper:
					log.Println("records: events - external termination")
					break scan
				case <-idxops:
					break ops
				case err := <-ch:
					if err != nil {
						panic(err)
					} else {
						recordsScan(idxFile, records, dropper)
						break scan
					}
				}
			}
		}

		close(records)
		log.Println("records: records closed")
	}()
	return
}
