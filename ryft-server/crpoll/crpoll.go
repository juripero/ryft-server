package crpoll

import (
	"os"
	"time"
)

var Interval = 50 * time.Millisecond

func sleep(s chan error) (err error) {
	select {
	case <-time.After(Interval):
		return
	case err = <-s:
		return
	}
}

func OpenFile(file string, s chan error) (f *os.File, err error) {
	for {
		if f, err = os.Open(file); err != nil {
			if os.IsNotExist(err) {
				if err = sleep(s); err != nil {
					return
				}
				continue
			}
		} else {
			return
		}
	}
}
