package datapoll

import (
	"os"
	"time"
)

var PollingInterval = time.Millisecond * 50

func Next(res *os.File, length uint16) (result []byte) {
	var total uint16 = 0
	for total < length {
		data := make([]byte, length-total)
		n, _ := res.Read(data)
		if n != 0 {
			result = append(result, data...)
			total = total + uint16(n)
		} else {
			time.Sleep(PollingInterval)
		}
	}
	return
}
