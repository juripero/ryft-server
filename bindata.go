// Code generated by go-bindata.
// sources:
// index.html
// swagger.json
// DO NOT EDIT!

package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _indexHtml = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xec\x5b\x7f\x93\xda\x36\xf3\xff\x3f\xaf\x42\x75\x26\x49\x93\xe1\x87\x6d\x38\x7e\xdd\x1d\xdf\xe1\x80\xb4\x37\x03\x5c\x72\x70\x4d\x32\x6d\xa7\x23\x6c\x01\x6a\x8c\xed\xca\xf2\x71\x5c\x26\xef\xfd\xbb\x92\x6d\x6c\x83\x0d\x34\x79\xb8\x79\xda\x79\xcc\xa4\x07\xd2\x6a\xa5\xcf\xee\x6a\xf7\x23\x79\x7a\xf1\x43\xef\xa6\x3b\xf9\xf4\xae\x8f\x16\x7c\x69\xb5\x9f\x5d\x04\x7f\x10\x3c\x17\x0b\x82\xcd\xe0\x2b\xfc\x58\x12\x8e\x91\xb1\xc0\xcc\x23\xfc\x52\xf1\xf9\xac\xd8\x50\xca\x51\x2f\xf4\x73\xca\x2d\xd2\xa6\xb6\x49\x1e\x2e\xca\xc1\x8f\x68\xa8\xc7\xd7\xf1\x2f\xa1\xbf\x80\xa6\x3e\xe7\x8e\x5d\x40\xd4\x76\x7d\x5e\x40\x1e\xb1\x88\x01\x7f\x39\x79\xe0\x98\x11\x8c\xbe\x6c\x14\x23\x34\x73\x6c\x5e\x9c\xe1\x25\xb5\xd6\x2d\xf4\x33\xb1\xee\x09\xa7\x06\x2e\x8e\x88\x4f\x0a\x1d\x46\xb1\x55\x50\x46\x74\x39\xf5\x3d\x34\xc6\xb6\x87\x06\x4a\xc1\x83\xbf\x45\x8f\x30\x3a\x3b\x0f\xf5\x7c\x0d\xff\x4e\x1d\x73\x9d\x52\x6e\x38\x96\xc3\x5a\xe8\x79\x55\x3e\xdb\xe2\xcf\x57\x0c\xbb\x29\xf9\x25\x66\x73\x6a\xb7\x90\x8a\xb0\xcf\x9d\xf3\x54\xcf\x43\x71\x45\x4d\xbe\x68\x21\x4d\x53\x55\xf7\x21\xd5\x49\xed\xa8\xb3\x59\xdb\xea\x73\xee\x09\x9b\x59\xce\xaa\x85\x16\xd4\x34\x89\x9d\xec\x73\xb1\x69\x52\x7b\x2e\xe6\x13\x1f\x7d\x6b\x68\xa8\xb2\x71\xf6\x62\x67\xe5\x2e\x9e\x93\xa2\xf4\x43\xc6\xfa\x8b\xdc\x71\x5b\xa8\x52\x89\xb5\x6d\xc6\x19\x60\x6d\x62\xf3\xd4\xa0\x29\x36\x3e\xcf\x99\xe3\xdb\x66\x31\xb2\xd7\x5b\xf9\x24\xd7\x32\x75\x98\x49\xa0\x4b\x73\x1f\x90\xe7\x58\xd4\x44\xcf\xbb\xf2\xd9\x15\x2a\x32\x6c\x52\xdf\x6b\x21\x1d\x64\x13\xff\x92\x92\x60\x11\xcc\x5b\xc8\x76\x6c\x72\xbe\x1d\x0c\x1e\x7d\x24\x30\x51\x25\x3d\xc2\xa2\x36\x29\x2e\x08\x9d\x2f\x60\x9c\xd6\x48\x77\x96\xdf\x44\x9e\x13\x36\x0c\x8c\x59\x7f\x71\xfe\xa6\x9c\x65\x6c\x29\x52\x79\x91\x1c\x0f\x3e\x12\x51\x67\x15\xb1\x45\xe7\xa0\x05\x0c\x98\x6b\xba\x36\x32\xe9\x7d\xca\x80\x26\xf5\x5c\x0b\xaf\xd3\x70\x76\xc6\xd1\xe5\xfc\x90\xd9\xfb\x3d\xf1\x39\xa1\xd9\x13\x51\x5c\xdf\x0e\xe2\x94\x7d\x72\x51\xe0\x16\x38\xe2\x73\xe6\x26\x53\xd5\xba\xde\xa9\x24\x55\x8a\x0d\x5f\x34\x89\xe1\x30\xcc\xa9\x63\x1f\x30\x50\xa0\xba\xb5\x10\x3b\x26\x35\xc1\x8e\x1a\xb0\x19\x61\x22\x20\x72\x75\xb9\x25\x93\x78\x06\xa3\xae\x18\x70\xc8\xec\xbd\xab\x7e\xe7\x6d\xf3\x80\xd9\xaf\xce\xba\xfd\x7e\x7d\x37\x58\x45\xf6\x6b\x21\xca\x21\x72\x8c\xfc\x78\x3d\xdb\xf6\x43\x10\xad\x5a\x10\xad\x8d\x1c\x4f\xd4\xf7\x38\xc2\xb7\x72\x32\x97\xf8\x54\xf6\xb9\x10\x46\x5a\x34\x35\x38\xce\x43\x59\x8b\xd4\xf7\xe8\x5a\x68\x29\x45\xe5\x37\x61\x2c\x4e\x1d\xa8\x00\xcb\xb4\x01\xe5\x93\xde\x93\x91\x07\x74\xf9\xe4\x64\x02\x9d\x2c\x77\x7a\x56\xa1\x61\x6d\x87\x2d\xb1\x95\x6f\xf7\xf4\x50\x19\x48\xe1\x16\x37\x60\xf9\x84\x65\xe1\x55\x33\x7c\x11\x03\xaa\xee\xb3\x86\x9e\xb9\x31\x2a\xf2\xc9\xcb\x73\xa5\xb3\xef\xc0\xd7\xc8\x8e\xab\x4a\x98\x05\x33\xb2\x44\xbe\x67\xbe\x19\x75\x25\x13\x75\xf3\x6d\x4d\xd7\x33\x36\x4c\x80\x7a\x0f\xe6\xa9\x63\x99\x99\x71\x18\x80\xd2\xf6\xc5\x76\xba\x9c\x1f\x65\xab\x86\xd0\x9a\x5d\x95\xbf\x7f\xf3\x1d\xa9\x6f\x6b\x3f\x7e\xcf\xd2\x38\x9e\xa6\x48\x41\x44\x5a\x54\x35\x51\xf1\xa4\x50\x11\x6a\x96\xe3\x83\x65\x66\xf4\x81\x98\x19\xa1\x02\x9e\xb4\xb0\xeb\x81\xbb\xa2\x6f\xbb\xf6\x0b\xb8\x46\x32\xbf\xe5\xac\x88\x9b\x85\x9d\xa6\x45\x3a\x33\xe7\x06\x68\xef\x4a\x7c\xce\x77\x83\x6c\x6a\x41\x36\xcf\x09\x31\xb5\xd4\x4c\x07\xd9\xc6\x78\xd5\xb4\x7f\x92\x59\xc1\x22\x33\x7e\x08\xc8\x22\xa7\x9e\x04\xe5\x0d\x31\xe2\x12\xcc\x11\x54\x1f\x30\x9a\x0c\x88\xe7\x6f\x35\xf1\x39\x2a\xe0\x0f\xd3\x86\xb0\x34\xa9\xfb\x8b\xa9\x87\x97\x6e\x61\x67\x77\x30\x72\x04\xf3\xcb\xa4\x20\x31\x47\xd1\x7b\x57\x8d\x2c\x81\xb0\x12\x4a\x97\x65\x74\x47\x71\x08\x5e\xdd\xfc\xcb\xdb\x28\xf5\x4c\x1e\xdc\xac\xef\xf2\xe0\x24\xdc\x2f\x31\xbb\x91\xd1\x2e\x7e\x86\x9b\xbf\x22\xc9\x4e\xc4\xd3\xa6\x96\x03\x51\x83\x56\x0b\xca\x49\xd1\x73\xb1\x01\xcb\x06\xd3\x44\xca\x63\xc6\x2e\x0f\x01\xbb\x39\x06\xac\x08\xa3\xec\xfc\x63\x8c\xd2\x75\x7c\x46\x81\xc5\x8c\xc8\x4a\x29\x84\x3f\x0a\x4b\xc7\x76\xe4\x6c\xf9\xbe\x0e\xce\x4b\x31\xaa\xf0\xe0\x14\x47\xe0\xc1\x13\x94\x32\xf0\x0d\x6a\xe2\x9f\x18\x06\x8e\xa4\x84\x27\xa8\xcd\xb1\x6a\xcf\xd9\x69\x33\x47\x69\xce\x08\xb1\x73\x18\x5e\xa3\xa1\xe6\xa7\x9e\xd2\xd4\xf2\x49\xce\x40\x31\x34\x7f\x60\x70\x34\xfa\x06\x52\x5d\x12\x1b\x36\x6d\x8d\xe0\x64\xb1\x77\x23\x97\x98\x88\x8a\xac\x61\xb2\xe3\x40\x02\x00\x6e\xe9\x18\xd9\xf9\x3f\xbf\x2e\x31\x22\x46\x15\xd2\x2d\x10\x17\xdf\x62\xe7\x60\x19\x16\xcd\x5d\x87\x9e\xac\x1b\x9b\xe1\x1e\x35\xc9\x14\xb3\xbf\x9f\xbd\xfe\x23\x87\xc1\x8a\x38\x74\xc5\xff\x32\x0e\x83\x49\x97\x89\x67\x93\x0e\x44\x75\x49\xd7\xc1\xbd\x07\x0b\xf1\xec\x39\xd0\x89\x27\xcc\x28\x7a\xc6\xc9\x3a\xb2\xd2\x56\x89\xb7\xa8\x17\x92\xfd\x22\x5f\xbb\x64\x77\xc6\xb0\x1c\x46\xd5\x2b\x83\x49\x17\x05\xc0\xd6\xce\x29\x3f\x59\x47\x8f\x2d\xf6\x89\x45\xe2\xcc\x08\xaa\xc9\xe7\x3c\x6b\x3f\x51\x5b\xd2\xa2\x20\x09\xe6\xb0\xb3\xed\xd3\xfa\x3e\x7a\x96\x22\x59\x7a\x35\x8f\xfb\xec\x64\xfb\xe3\x0e\x87\x49\xa4\x19\x47\xc3\xac\x0a\xa6\x8b\x4f\x06\x59\x78\x5e\xed\x36\xab\x57\x95\x7d\x53\xc0\x9e\xc2\x06\xa7\xf7\x64\xcb\xac\x47\x5d\x91\xfc\x9d\x69\x0e\xa5\x8f\xf0\xcb\x45\x39\x71\xbb\x76\x51\x8e\xaf\xec\x2e\xc4\x55\xd7\xe6\x0a\x4e\x5c\x48\x50\xf3\x52\x09\x13\x84\x22\x2e\xfa\xb4\x36\xba\x5d\x43\x6e\x1c\x13\x26\x8c\x06\x83\x35\xd1\xac\xb7\xc7\x04\x33\x63\x81\x88\x6d\xba\x0e\x85\x6c\x52\xf6\x82\x06\x17\x33\xbc\x24\x70\x26\xf2\x50\x0b\xa4\x75\x90\x96\x89\x46\xcc\x72\xc1\x13\x53\x73\x96\xbc\x18\x5c\x20\xb9\xc4\x4b\x25\xac\xbc\x22\xbc\x95\xf6\x90\xf0\x85\x63\x5e\x94\xf9\xe2\xb0\xec\xb5\xa8\x7a\x48\xec\xaa\x03\xf2\xb2\x8a\x2b\xed\x3b\x46\x0f\x08\x56\x03\xc1\x5e\x7c\x05\x10\x0f\x80\x6f\x72\xfd\xa2\x25\x81\x29\x65\xcf\x24\x42\xf1\xd3\x6c\x5f\xe0\xf6\x5f\x3e\x61\xeb\x8b\x32\x6e\xc3\x48\x73\xbb\xdf\xe3\x0c\xc2\x3c\xab\xe7\xa7\xfe\x24\x32\xf1\xff\x49\x15\x97\x5f\xde\xdf\xf5\x6f\x3f\x7d\xcd\x12\x1e\x4b\x35\x40\x31\x45\x0a\x76\x89\x41\x67\xeb\xe0\x37\x30\x8e\xc0\x4b\x00\x08\x5c\x44\x71\x09\xdd\x92\xbf\x7c\xca\x88\x09\xf4\x5d\x14\xa5\xc8\x7b\x49\xb5\x11\xd6\x1d\xa7\x49\x40\x62\x9c\xb7\x0b\x28\x0f\xce\x5e\x30\x2f\x81\x7f\x9d\x4b\x85\x97\x5f\xde\x5e\x0f\xfa\x5f\x77\xc7\x06\x6e\x36\x31\xc7\x80\x05\xfc\xed\xa0\x69\x84\x8a\x98\x25\xd4\x75\x96\x4b\xd1\x23\x80\x70\x40\x25\x12\x2e\x72\x66\x12\x9d\x87\x1c\x06\x09\x8c\x01\x19\x72\x80\x4a\x79\xa5\x58\xfb\x01\x88\xfe\xe3\x23\xa4\x27\x2f\x07\xa6\x0f\xf1\xdf\xf8\x1e\x94\x41\x43\x34\xc9\xe5\x97\x5f\x3a\x83\xbb\x2c\xec\xe3\xc0\x97\xd2\x91\x42\x7c\x1d\xb9\x13\xb2\x32\xc7\xb6\x41\xd0\xaf\x6a\xa9\xa4\x9f\x9d\xfd\x8e\x8e\xc6\x66\x9c\xcc\x77\xb2\xc1\xf0\x2e\x39\xf3\xc9\xae\x9e\x2e\xf6\x84\xdf\x6c\x8f\xca\x44\x39\xb3\xf0\xbc\x84\x7a\x64\x86\x7d\x8b\xa3\x57\x33\x6c\x79\xe4\xd5\xf1\x0e\x12\xf7\x1b\xfc\x74\x40\xdc\xe5\x79\x30\x05\xb4\xdc\xdc\x0e\x3b\x93\x0c\xe7\xbc\x8b\xb6\x0e\xd4\x39\x16\xec\x35\x40\x6e\x88\x20\x0c\x66\x29\x21\x94\x74\x20\x8d\x03\x39\xd0\x8d\x5e\x3d\x2c\xad\x57\x22\x46\x5f\x31\xbc\x7a\xf5\x63\x68\x8c\xd7\x47\x5b\xc1\xf3\x19\x83\x5c\x6c\x4a\xb8\x79\x81\xaa\xd5\xbe\xdb\xa7\x72\x1e\x5f\xce\x93\x1f\xab\xb1\x39\x12\x69\x08\x36\x9d\xc4\x6e\xfb\xcb\x29\xf4\xc0\xbe\x14\x2f\x8b\xa0\x56\x8a\x72\x31\x25\x60\x07\x22\xfb\xc1\x1a\x10\xd5\x70\xf6\x40\x78\x16\xa8\x88\xda\xa4\xb2\x15\x05\x52\x39\x15\x2c\x93\xfb\xcc\x06\x03\xaf\x16\x40\xfb\x63\xa3\x46\x93\x31\x59\x0e\x10\xf5\x10\x18\x54\x32\x85\xa3\x03\x8a\x12\xcb\x3c\xed\xce\x08\x03\x0a\x7c\x1e\xf6\x8b\x19\x85\x40\x7f\xd0\x1b\x97\x4a\xa5\x7c\x9b\xee\x98\xd4\x26\xc4\x04\x2b\x7c\x26\x6b\x0f\x2c\x00\x66\xf1\x20\x6e\x92\xa9\x7d\x33\xd5\xd1\xb1\x64\x3b\xe6\xe9\xb2\xba\x6c\x90\x33\x7c\x43\xfc\xc8\x71\xc0\x92\x7c\x20\x1d\x22\xdd\x55\x7f\x8f\xf3\x46\xb5\x80\xe8\x2c\x90\xb8\x54\x91\xb7\xf6\x38\x59\x06\xd1\xe2\x43\xb6\x31\x43\xa9\x7b\x0c\x07\xcc\x1d\x4b\xc0\x7f\x83\xea\x0d\x5f\x02\xc2\x72\xf1\x43\xb1\x88\x04\xd7\xe9\x3f\xc0\x82\x65\x95\x83\x1f\xa8\x58\x14\x0c\xa8\xd2\x1e\x39\x3c\xde\xe2\x0c\x6c\x4d\xa0\xd4\x90\x40\x14\x24\x2b\x20\xe5\x82\x29\xd1\x82\x91\xd9\xa5\x92\xb6\xcc\x8f\xb7\x9d\x0f\x7f\x4c\xfa\x1f\x27\x2f\x74\xb5\x7b\x33\x9a\x74\xae\x47\x63\xf8\xfa\x42\xd7\x35\xf1\x9f\xd7\x2f\x03\x73\xb9\xd8\x83\x14\x39\x87\xdd\x51\xe2\x0f\xfc\x65\x72\xe3\x69\xea\xcb\xb8\x62\xa8\x8a\x40\x92\x33\x05\x8a\x26\x40\x8a\xa6\x2a\xaf\x13\xde\xd8\x52\xbf\xbd\xb7\x61\x8a\x74\x61\x52\x21\x20\x10\x44\x84\x2b\x99\x8f\xb8\x2c\x69\xff\x2a\x2d\x18\xd3\x50\xe5\x0f\xf9\x76\x57\x69\xa5\xa8\x29\xb4\x8b\x19\xa1\x55\x29\x33\xa0\x93\x40\xd1\xcb\xe9\xc9\x95\x42\x5a\xdc\x99\xcd\xa0\xbc\xc3\x00\xbd\xbe\xd5\x63\xc1\x20\xbe\x10\x3d\xfa\x56\xcf\x66\xa1\xd0\xa9\x6e\xba\xbe\xc6\x52\x8a\xc8\xb6\x62\x11\x9f\x3e\xac\xe6\x77\xba\xe6\x9a\x3f\xcd\xbd\xeb\x5e\x7f\x35\x98\x74\x1e\x06\x93\xfe\xd9\xf0\xcf\xb9\x37\xec\x5c\x5e\x2a\xcf\x12\x23\x4f\x08\xae\x5a\x39\x01\xb8\xe1\xe4\xf3\xfa\xa6\xbb\x5a\x0d\x27\x7d\x3e\x7c\x14\xe0\xee\xb4\xd1\x58\x7d\x18\xfe\xd9\x5f\x0f\x9e\x10\x9c\xa6\x36\x4e\x80\x6e\x30\xb9\x3e\x13\xae\x1a\xf5\xee\xbc\x04\xba\xf5\xf0\xf1\xfa\xf1\x49\xd1\xd5\x4e\x11\x98\x83\xc9\x70\x15\x04\x62\x27\x89\xee\x71\xd4\x1b\xaa\x4f\x89\x4e\xaf\x54\x4f\x13\x99\xda\x90\xee\x44\xa6\x3a\x9a\xbc\xd7\x9e\x12\x5d\xa5\x7a\x0a\x74\xe0\x37\x2d\x88\xcc\xf7\x49\xdf\x69\xa3\x3f\xef\xf4\xa7\x44\x57\xad\x9f\x66\xdf\xf5\xd5\x00\xdd\xe7\x24\x3a\x7d\xf4\xf8\xa9\xf2\x94\xe8\xce\x6a\xcd\xd3\xa0\x5b\x4b\x74\x93\x14\x3a\x91\x33\x9f\x34\xab\xd4\x6a\xa7\xa8\x08\xff\x2d\xe8\xea\x75\xf5\x5f\x8c\xae\xd1\x3c\x25\xba\x9b\xf4\xbe\x83\xdf\xc6\x93\x66\x95\x66\xad\x76\x9a\x6a\xae\x05\xe8\xee\xb6\xea\xdd\x75\xb2\x22\x3c\xfb\xfd\xa2\x2c\xe9\xa6\x24\xde\xe3\x6f\x26\xdd\xfd\xee\xcd\x6d\xaf\x74\xdd\x03\xaa\xdd\x7f\x7f\xd7\x19\x6c\x38\xb7\x5a\xa9\x6a\x8d\x4a\x82\x79\xbf\x29\xb9\x06\xa3\x4b\xb2\x87\x73\xbf\x8c\x8f\x55\x99\xf4\x3b\x9a\x0c\x05\x53\x09\xf6\x1d\x4c\x93\xe2\xe0\x9b\x89\x0e\xb3\xef\xad\x63\x63\x4c\xc6\x03\x2a\x1e\x38\x5a\xe9\x30\x38\xff\x09\x8f\x29\xf2\x12\x25\x74\xa8\x72\x45\xb0\x6c\x54\x2b\x5a\x65\xd3\x26\x2e\xef\x65\x63\x4d\xff\xf8\x11\x8d\xd1\x78\x82\x06\x9d\x0f\xb7\xfd\x51\xb7\x8f\x3a\xbf\xf4\x23\x39\x71\x57\x33\x92\x47\x76\x21\xfc\xf3\x27\x5d\xaf\xd4\xea\x1b\x2d\xe2\xca\xcd\xb7\x29\x5f\x77\x18\x91\x4e\xad\xea\x51\x57\x0f\x73\x19\x6c\x6a\xb5\xac\x9d\x95\x75\x55\x3b\x43\x9a\xd6\x3a\x6b\xb6\x54\x15\xbd\x1b\x6e\xa4\xe2\x0b\x56\x21\xdc\xbb\x19\xf6\xc7\x93\xeb\x2e\xba\xea\x4c\x26\x70\x7e\x44\xe3\xeb\xe1\xbb\xc1\x66\x31\x3d\x2a\x0e\xa0\x46\x00\x46\xdd\xac\xa2\xe7\x2c\x01\x36\x35\x44\xb3\xb8\x71\x8a\xda\xdf\x5e\x5d\x77\xe1\x3c\x28\xa5\x1b\x57\x51\xeb\x75\x4f\x34\x6c\x1c\x12\xb5\xde\x75\x6f\x83\xe5\x36\x6a\x51\xdb\x00\x73\xca\xfd\x40\x41\x55\x2b\xd5\x1b\x5a\xb3\xa6\xd5\x1a\x8d\x4d\xbf\x63\xe0\x68\xe5\xbf\x29\x3f\x26\x45\x0a\xa8\xd8\xa8\x97\x6a\x9a\xda\x6c\x54\xeb\xea\xd9\xeb\xdf\x94\xed\x41\x5b\xc8\xc7\x93\xdb\x7e\x7f\x12\x0b\xd9\xf3\xcd\xd4\x29\x4d\x91\xc4\x3b\x08\x1c\xcc\xd6\x93\xb5\x2b\x65\x42\x7b\x45\xbd\x77\xae\x29\xae\x40\x6f\xec\xd0\x03\xba\x1e\x7a\x40\x6f\x55\xeb\x2d\x2d\xe9\x81\x0f\x98\x99\x42\x4a\x57\xa3\x96\x8f\x5d\xc7\x61\x10\x87\xa1\x03\x35\xad\xa1\xe9\xb5\x8d\xa1\x3e\x6d\xf5\x36\x6a\x95\x66\x6d\xb3\xac\x4f\xb0\x0f\x02\x6d\xda\xa6\x2d\x23\x1d\xed\xa6\x22\x63\x41\x0d\x3c\x77\xc2\x1d\x91\x48\x45\x71\x1a\x4a\xa4\xd8\x38\x01\xd5\x9a\x89\xaa\x99\x91\x7d\xbe\x3e\xfb\x1a\xe5\x0f\xf9\xee\xa2\x2b\x2f\x0f\xa2\x57\x17\xe1\x5b\x0a\x79\xd8\x77\xdb\x13\x27\xbc\x5b\x20\x16\x59\xda\xdc\x93\x37\x07\x65\x23\x35\xa2\x84\x3e\x38\xec\xb3\x87\x56\x94\x2f\x10\xb6\xac\xe4\x4b\x0f\x3c\x75\xee\x49\x09\x89\x7d\x29\x6f\x0a\xfe\xf7\xfa\x23\xfa\xf9\xdd\xaf\x3f\xa4\x13\xfe\x25\x6f\x3f\x32\xb0\xfc\xfb\x5e\x7e\x1c\x03\xf2\x1f\xf7\xee\xe3\x68\x50\x4f\xf3\xea\xe3\x5b\x2f\x6a\x8f\x86\xf1\x0f\xb8\xa7\x05\xda\x17\xa4\xf4\xc3\x84\x30\x09\x3b\xa4\x68\x3b\x57\xb0\x58\xf0\xc0\x9b\xdb\xbc\xee\x69\x06\x4d\x0c\x08\x60\x86\xf2\xc4\xe5\x2b\x56\x62\xa5\x89\xe6\x69\x26\x1d\x4c\x90\x3b\x97\x91\xb6\xa0\x2a\xaa\xaa\x40\x03\x13\x80\xcb\x26\xbd\x8f\x6c\x12\x19\x23\xf8\x7f\x72\xfe\x3f\x00\x00\xff\xff\x88\x01\x66\x54\xab\x33\x00\x00")

func indexHtmlBytes() ([]byte, error) {
	return bindataRead(
		_indexHtml,
		"index.html",
	)
}

func indexHtml() (*asset, error) {
	bytes, err := indexHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "index.html", size: 13227, mode: os.FileMode(436), modTime: time.Unix(1451919984, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _swaggerJson = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xe4\x57\xdd\x4f\xe3\x46\x10\x7f\x4e\xfe\x8a\x91\x9f\x5a\x29\x17\x12\xd4\x70\x94\x37\x84\x40\x45\xa7\x42\x4b\x90\x5a\xa9\x77\xaa\x16\x7b\x12\xef\x9d\xbd\x6b\xf6\x83\x84\x43\xf9\xdf\x3b\xbb\x5e\x07\x7f\x85\x43\x85\x56\xa2\x7d\x49\xec\xd9\xdf\xec\xcc\x6f\xbe\x76\xfd\x30\x1c\x44\xb1\x14\xda\xe6\xa8\xa3\x23\xf8\x63\x38\x18\x44\xac\x28\x32\x1e\x33\xc3\xa5\xd8\xfb\xac\xa5\x88\x46\x6d\xe9\x3a\xcf\xa2\xe1\xe0\x13\xc9\xa3\x42\xc9\xc4\xc6\xfd\xca\xb9\x5e\x16\x2c\xfe\xd2\xd5\xf7\xbb\x86\x0d\x74\x9c\x62\xcd\x78\x6a\x4c\x51\x2a\xb8\x27\xbd\x45\xad\xd8\x72\x89\x8a\x50\xd1\xfe\x78\xe2\x00\x11\x17\x0b\x49\xef\x0f\x0e\x9b\xa0\x8e\x15\x2f\xdc\xe6\x0e\x62\x52\x84\xc2\xaa\x42\x6a\x04\xb9\x00\x93\x72\x0d\x35\xfb\x40\xaf\x46\x42\xac\x90\x19\x84\xab\xd3\xf9\x35\x50\x0c\x04\xc6\x46\x2a\xb7\xe0\xd4\xaf\xee\x17\x06\x2e\x05\x42\xca\x54\xb2\x62\x0a\x4b\xa7\x0c\x37\x19\x3a\x13\x7e\xdd\xa9\x8e\xcb\x85\x3b\x54\x3a\x58\x9f\x92\x83\xc3\xc1\xc6\xf9\x98\x4a\x6d\xbc\xe8\xc7\xfd\xf1\xf4\xe0\x70\x3c\x7b\x3f\x9e\x4e\xa6\x47\x87\xef\x0f\x66\x9e\x43\xc1\x4c\xaa\x2b\x12\x7b\xb1\xb4\xc2\x84\xb7\x41\xb4\xc4\xed\x73\x87\xe0\xa9\x48\x0a\xc9\x85\x81\x85\xf3\x98\xdc\x2d\x55\x47\x25\xd8\xb0\x65\x15\x4f\xf7\x5a\xae\xf9\x97\x4f\x01\x41\xf9\xce\x99\xba\x77\x5b\x9d\xd4\x35\x65\x81\xca\x87\xe8\x3c\x71\x6b\x8d\x5d\x0b\xa6\x58\x8e\x86\x78\xba\xbd\x1f\xc2\xe6\xe6\xbe\xf0\xf1\xd0\x46\x71\xb1\x0c\xd8\x41\xb4\x7e\xb7\x94\xef\x04\xe1\xdd\xda\xaf\x16\xc9\x56\xb5\xd4\xa2\x32\x47\xa6\xe2\x14\x6e\x1d\x66\xe4\xf9\xe0\x9a\xe5\x45\x86\x47\xf0\x1d\x5c\x1d\xff\xf6\xe7\xf5\xe9\xef\xd7\x70\x72\x79\x71\x7d\x7c\x7e\x31\x87\x8f\x91\xe0\xcb\xd4\x7c\x8c\xe0\xfb\xed\x8e\x95\x9d\xdb\x86\x1d\x2e\xba\x32\x85\xb7\x96\x2b\x74\xe4\x8c\xb2\xe8\xa5\x9b\x11\xb4\xc9\x30\xa5\x58\x6d\x23\x83\xb9\xde\xa6\xa2\xcb\xb9\x14\x6f\x7a\xa9\x9f\xf1\x8c\x6a\x7b\x17\x75\x69\x55\x8c\xb0\x68\x60\x2a\xcd\xa6\xf4\x05\x64\xa8\x4c\xd0\xb5\x4e\xa5\x46\x21\xce\x99\xaf\x4a\x4b\x4b\x87\xfd\x29\x3b\xb3\x5f\xbf\x72\x81\x7a\xa7\xef\xe7\xda\xd7\xdd\xa2\xc2\x95\x7d\x86\xa0\x7d\x3a\xc7\xf0\x33\x32\x6d\xc9\x39\x60\x25\x30\x67\x6b\x9e\xdb\x1c\x7e\x62\x79\x4e\x41\x83\x84\x6b\xc3\x44\x8c\xe3\x2e\xf1\x8e\xe9\x3a\xf9\x5d\x2c\x6f\xa4\xcc\x90\x89\x7e\x36\x27\x4c\xe3\x1c\x85\xe6\x86\xdf\xe1\x2e\x46\x0e\x44\xee\x07\x14\x2c\x32\xb6\xec\xf8\x16\x3f\xcf\xa9\x40\x96\x00\x3f\x54\x70\x22\x1d\x44\x93\xd1\x6b\x24\xe8\x42\x26\xbb\x0b\xeb\x38\xf6\x14\x3c\x06\xea\x0d\xfe\xc8\x44\x34\xf4\xbb\x64\xaa\x49\xa1\x50\x17\x74\x3e\x60\xad\x01\xa2\xfd\xc9\xa4\xde\x0e\x7e\x82\xb3\x9a\x64\xd7\x54\xa8\x99\x27\x82\xa4\xa5\x43\xef\x0c\x36\xc3\xda\x9f\xff\x75\x3f\xbe\xa5\xa2\xbd\xb2\xa4\xfe\xde\x5c\x0c\xba\xfd\x83\x31\x2c\x36\x26\x63\x6b\x02\x36\xf5\x9f\x1e\x81\xff\xa9\xa9\xf1\xd4\x3c\x3f\x2b\x0b\xf4\x1b\xfe\x79\x10\x50\xc8\x34\x2a\xf0\xaa\x1d\x67\x9b\xfb\xbc\xb9\x36\x7f\x56\xac\x38\x66\x49\x5f\x9e\x1a\xe2\xe7\x58\x79\xbd\xf2\xfa\x80\xf7\x5d\x8f\xbe\xd4\x85\x2f\x65\xfd\x66\x4e\xfc\xc7\x49\x7d\x30\x9b\xbd\x7c\x58\x4f\x0f\xfa\xe3\x31\xb7\x4a\xd1\x18\x4e\xea\x01\x6b\x47\xa5\xc0\x98\x53\x59\x94\xc7\xa5\xb0\xf9\x0d\xb5\x0d\x1d\xaa\x31\x5d\x40\x59\xec\x66\x0e\xdc\x20\x59\xc3\x70\x9c\xd2\xfc\x04\x26\xe8\x7c\x5d\xd0\x5a\x4d\x66\x52\xea\xba\x15\xcf\x32\x82\x83\x42\x63\x95\xa0\x53\x78\x95\xa2\xf0\x20\x2e\x0a\x6b\x40\x07\x63\xa4\x48\xbc\xdc\x7d\x58\xb1\x15\x18\x5c\x77\x0f\x0a\xdd\xe3\xfa\xf3\xce\xbe\xfd\xd9\xec\x9f\x3a\xfd\xde\xcc\xf5\xe4\xff\x75\x13\xe8\x6d\xf4\xab\xa0\x77\xf9\xc1\xd5\x19\x0b\x29\xa0\xd2\xd4\x36\x33\x54\x8f\x54\xc5\xe0\x67\x1b\x55\xad\x71\x35\x0d\x98\x91\x6b\xc2\xf8\xa4\xdd\x31\xc5\xa5\xd5\x40\xd3\xc6\xc6\x54\xcc\xdb\xb9\xfe\xe4\xbd\xa3\x31\x2b\x7b\xa6\x65\x0d\x2a\x6f\x3e\xd3\x87\xdf\x23\x96\xbe\x54\x93\x84\x3b\x0a\x2c\xfb\x45\xb9\xeb\x80\xe1\xd8\xd4\xed\x28\x6f\x57\x36\xc3\xd6\xc3\x66\xd8\x7f\xb3\x09\xdf\x87\x09\x2e\xa8\x1a\x9c\x31\x6f\xc1\xcb\xda\x71\x8e\x9a\x9f\x85\xed\xe3\xcc\x2d\x56\x21\xae\xfa\x5d\x43\xb8\x5f\xf9\xe9\xba\x4d\x66\x79\x9f\x6a\x5d\xa7\xfe\xbd\x9c\xb5\x13\xd6\x9b\xad\x76\xaa\x76\xe4\xe9\x5b\x49\xea\xcf\x50\x4f\x16\x86\x9b\xe1\x5f\x01\x00\x00\xff\xff\x75\x4b\x73\x6e\x04\x11\x00\x00")

func swaggerJsonBytes() ([]byte, error) {
	return bindataRead(
		_swaggerJson,
		"swagger.json",
	)
}

func swaggerJson() (*asset, error) {
	bytes, err := swaggerJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "swagger.json", size: 4356, mode: os.FileMode(420), modTime: time.Unix(1452098645, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"index.html": indexHtml,
	"swagger.json": swaggerJson,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"index.html": &bintree{indexHtml, map[string]*bintree{}},
	"swagger.json": &bintree{swaggerJson, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

