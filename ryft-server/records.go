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
	//records = make(chan IdxRecord, 64)
	records = make(chan IdxRecord, 1024) // for debugging reasons
	go func() {
		log.Printf("records: start records scanner")
	scan:
		for {
			log.Println("records: start iteration scan loop")
			for {
				log.Println("records: start iteration readline loop")
				var line string
				n, _ := fmt.Fscanln(idxFile, &line)

				if n == 0 {
					log.Println("records: nothing for read -> signals loop")
					break // waiting for write event
				}

				log.Printf("records: received line for parsing: %s", line)

				r, err := NewIdxRecord(line)
				if err != nil {
					log.Printf("records: record parsing error '%s': %s", line, err.Error())
				}
				records <- r
				log.Printf("records: sent: %+v", r)
			}

		ops:
			for {
				log.Println("records: start iteration signals loop")
				select {
				case op := <-idxops:
					log.Printf("records: received op %+v", op)
					break ops
				case err := <-ch:
					if err != nil {
						log.Printf("records: received error from progress")
						panic(err)
					} else {
						log.Printf("records: received normal completion from progress")
						break scan
					}
					// default:
					// 	log.Println("records: no external signals -> next iteration of scan loop...")
					// 	break ops
				}
			}
		}

		close(records)
		log.Println("records: closed")
	}()
	return
}
