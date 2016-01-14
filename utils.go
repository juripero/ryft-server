package main

import (
	"fmt"
	"log"
	"net/url"
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

func createClusterUrl(params *UrlParams) string {
	if params.host == "" {
		log.Fatal("Host coudn't be empty" + params.host)
	}
	if params.Scheme == "" {
		params.Scheme = HTTP
		log.Println("Empty scheme. HTTP scheme will be used by default")
	}
	u := &url.URL{}
	u.Host = params.host
	u.Scheme = params.Scheme
	u.Path = params.Path
	q := u.Query()
	for k, v := range params.Params {
		q.Set(k, fmt.Sprintf("%v", v))
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func createFilesQuery(files []string) string {
	result := ""
	for i, v := range files {
		if i != len(files)-1 {
			result = fmt.Sprintf("%s%s,", result, v)
		} else {
			result = fmt.Sprintf("%s%s", result, v)
		}
	}
	return result
}
