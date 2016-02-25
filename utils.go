package main

import (
	"fmt"
	"log"
	"net"
)

type AddressMaker interface {
	SetHost()
}

type UrlParams struct {
	host   string
	Scheme string
	Path   string
	Params map[string]interface{}
}

func (params *UrlParams) SetHost(address string, port string) {
	if address == "" {
		log.Fatal("Couldn't parse emty url")
	}

	if port == "" {
		port = DefaultPort
		log.Println("Empty port. Port 8765 will be used by default")
	}
	params.host = fmt.Sprintf("%s%s%s", address, HostPortSep, port)
	log.Print(params.host)
}

const (
	HTTP        = "http"
	HTTPS       = "https"
	DefaultPort = "8765"
	HostPortSep = ":"
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
