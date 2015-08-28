package records

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var Capacity = 64
var PollingInterval = time.Millisecond * 50

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

func scan(f *os.File, drop chan struct{}, out chan IdxRecord) (err error) {
	var line string
	var r IdxRecord
	for err != nil {
		if n, e := fmt.Fscanln(f, &line); n == 0 {
			if e != nil {
				log.Printf("%s: scan err: %s", f.Name(), err.Error())
			}
			break
		}
		log.Printf("%s: RAW: %s", f.Name(), line)

		if r, err = NewIdxRecord(line); err != nil {
			log.Printf("%s: parsing err '%s': %s", f.Name(), line, err.Error())
			break
		}

		log.Printf("%s: sending %+v", f.Name(), r)
		select {
		case <-drop:
			err = fmt.Errorf("%s: external termination!", f.Name())
			break
		case out <- r:
			log.Printf("%s: sent %+v", f.Name(), r)
		}
	}

	return
}

func sleep(
	drop chan struct{}, s chan error,
	timeout, dropped, complete func() bool,
	completeWithError func(err error) bool,
) bool {
	select {
	case <-time.After(PollingInterval):
		return timeout()
	case <-drop:
		return dropped()
	case err := <-s:
		if err != nil {
			return completeWithError(err)
		} else {
			return complete()
		}
	}
}

func Poll(idx *os.File, s chan error) (records chan IdxRecord, drop chan struct{}) {
	records = make(chan IdxRecord, Capacity)
	drop = make(chan struct{}, 1)
	go func() {
		loop := true
		for loop {
			if err := scan(idx, drop, records); err != nil {
				log.Printf("%s: READ WITH ERR: %s", idx.Name(), err.Error())
				close(records)
				return
			}

			loop = sleep(
				drop, s,

				// Timeout
				func() bool {
					log.Printf("%s: TIMEOUT.", idx.Name())
					return true
				},

				// Dropping connection or another external reason to stop records reading
				func() bool {
					log.Printf("%s: DROPPED CONNECTION.", idx.Name())
					close(records)
					return false
				},

				// Search complete
				func() bool {
					log.Printf("%s: SEARCH COMPLETE.", idx.Name())
					scan(idx, drop, records) // Connection can be dropped or not.
					close(records)
					return false
				},

				// Search complete with error
				func(err error) bool {
					log.Printf("%s: API ERROR: %s", idx.Name(), err.Error())
					close(records)
					return false
				},
			)
		}
	}()
	return
}
