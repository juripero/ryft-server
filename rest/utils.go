package rest

import (
	"net"
)

// FIXME: review this function
func compareIP(inIP string) bool {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				if ipnet.IP.String() == inIP {
					return true
				}
			}
		}
	}
	return false
}
