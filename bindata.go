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

var _swaggerJson = []byte("\x1f\x8b\x08\x00\x00\x09\x6e\x88\x00\xff\xd4\x56\x5b\x6f\xdb\x36\x14\xfe\x2b\x07\xdc\x1e\x36\x40\x75\xec\xa0\x4d\x33\xbf\x05\x41\x82\x05\x43\x53\xcc\x0e\xb0\x01\x6b\x31\x30\xd4\x91\xc5\x55\x22\x15\xf2\x30\x76\x1a\xf8\xbf\xef\x90\x92\x1c\xf9\x52\x20\xd9\xc3\xd0\xe4\x45\x26\xbf\x73\xf9\xce\x95\x79\x14\xca\x1a\x1f\x6a\xf4\x62\xfa\x97\x90\x4d\x53\x69\x25\x49\x5b\x73\xf4\x8f\xb7\x46\x64\x5b\x57\xab\xba\x12\x9f\x33\xd1\x38\x9b\x07\xb5\xaf\x51\xfb\x45\x23\xd5\x97\x1d\xa5\x64\x87\xb5\xbc\x2a\xb1\x73\x53\x12\x35\x2c\x15\x3f\x3e\x41\x4b\xb9\x58\xa0\x13\x53\x71\x3c\x1a\x33\xa0\x4d\x61\xc5\xf4\x51\xe4\xe8\x95\xd3\x4d\x34\xc3\x18\x95\x08\x4d\x70\x8d\xf5\x08\xb6\x00\x2a\xb5\x87\x81\x23\xe0\x23\x59\x50\x0e\x25\x21\xcc\x2e\xe6\x37\xc0\xa1\x19\x54\x64\x5d\x04\xa2\xfa\xec\xa1\x20\xf8\x68\x10\x4a\xe9\xf2\xa5\x74\xc8\xce\x48\x53\x85\x6c\x3e\x61\x51\x6d\xc4\x97\xf7\xe8\x7c\xeb\x75\xc2\x8c\xd6\xcc\xd5\x7a\x8a\xa7\x5f\x8e\x47\x93\x93\xd3\xd1\xbb\xf7\xa3\xc9\x78\x32\x3d\x7d\x7f\xf2\x8e\xa5\x1b\x49\xa5\x8f\x84\x8f\x94\x0d\x86\xe2\xaf\x05\xd2\x7e\x04\x17\x26\x6f\xac\x36\x04\x45\xa4\xc4\x7c\x5a\x71\x4e\x40\xa8\x6b\xe9\x1e\x58\xe4\xbc\xbb\xb1\x0d\xba\x14\xd6\x55\xce\xb7\xbd\x5c\x23\x9d\xac\x91\x98\x1c\xa7\xf1\x51\xd0\x43\x13\x99\x7b\x72\xda\x2c\x18\x5e\xbd\x59\xd8\x37\x86\x25\xf8\xf2\xf7\x80\x6c\x30\xdb\x61\x30\x47\xe9\x54\x09\x77\x11\xcc\x12\x0d\x5c\xc9\xba\xa9\x70\x0a\x3f\xc1\xec\xec\x8f\xbf\x6f\x2e\xfe\xbc\x81\xf3\x8f\xd7\x37\x67\x57\xd7\x73\xf8\x24\x8c\x5e\x94\xf4\x49\xc0\xcf\x6c\xaa\xb3\x7c\xd7\x59\xd6\x66\x70\x70\x78\x17\xb4\x43\x26\x4b\x2e\xe0\x3a\xdb\x90\x93\xce\xc9\x24\x4d\x58\xa7\x1c\x6d\x93\x5e\x6f\xb3\xbe\xd4\x15\xb7\xc8\x1e\x6b\x1b\x9c\x42\x28\x3a\xb0\x93\xed\x8f\xcf\xe3\xc1\x69\xc7\xd8\x63\x99\xe0\xa8\x6b\x19\x8b\x19\xf8\xee\x74\x27\x6d\x97\xe1\xeb\x57\x6d\xd0\xef\x93\xb8\xf2\xa9\x64\x45\x2f\xd0\xf6\x20\x82\x4f\x29\x1d\xc1\x07\x94\x3e\xb0\x67\x90\xad\x60\x2d\x57\xba\x0e\x35\xfc\x2a\xeb\x9a\x23\x85\x5c\x7b\x92\x46\xe1\x68\x10\xc1\xc0\xd9\x20\x8a\x01\xeb\x5b\x6b\x2b\x94\x66\x87\xe4\xb9\xf4\x38\x47\xe3\x35\xe9\x7b\xdc\x23\x1a\x51\x66\xd5\xc1\x50\x54\x72\xf1\xe4\x52\x1d\xf0\xd5\x31\x15\xd3\xb7\x99\x60\xaa\xed\xef\x71\xf6\xc2\xcc\x5d\xdb\xfc\x40\xe9\xce\x54\x22\x91\x40\xe8\x9b\xbb\xd3\x30\x9d\xc6\x90\xce\xe7\x58\x42\xdf\xf0\x4e\xc2\xd4\x2d\xc7\xe3\x71\xfc\xfc\xe8\xb0\x60\xa1\x1f\x8e\x36\x60\x3b\x6a\x33\x3e\x8a\x35\xff\x65\xe2\xa8\x2d\xc4\xf3\x87\xaf\x93\x1f\x4e\xdf\xbc\xbf\xda\x1e\xbf\x8d\xe4\xe1\xf9\xfb\x5f\x8a\xf4\x9f\xc6\xe9\x37\x7c\x18\x0c\xcc\x97\xf6\xf4\x1d\x57\xff\xd5\xac\x8d\xc9\xc9\x4e\xfc\xf3\xe0\x1c\xc7\x97\xb7\x8b\x78\x87\x47\x83\x4a\x17\x1a\xdb\xbd\x60\x42\x7d\x8b\x2e\x6e\x0f\xc5\xaf\x90\x54\xb1\x9b\xe0\x16\xd9\x3c\x76\x7b\x83\x78\x43\x4b\xc3\x8b\xa4\x60\x6c\x70\x47\xa5\x24\x58\xea\xaa\x62\x71\x70\x48\xc1\x19\x5e\x37\xcb\x12\x4d\x12\xd2\xa6\x09\x04\xbe\x73\xc6\x8a\x1c\x41\x7c\x14\x9d\x5c\x02\xe1\x6a\x90\x7a\xbf\x45\xf6\x70\x01\x0e\x3e\x2a\x97\x6d\x0e\xbe\x95\xe8\x84\x02\x0f\x89\x67\xf7\x49\xe5\x29\xeb\xbd\xe6\x4b\xbc\x69\xac\xf2\xad\xca\x75\xe7\x17\xd8\xf8\x5e\x9e\xc1\xd7\xf4\xfc\x3c\x7f\x07\xb7\xce\x37\x4b\x78\xbd\xa3\xf8\xb4\xa2\xf7\xb6\xf1\x79\x0f\x45\xe1\xbe\x99\x7d\xdb\xe8\xfc\x8d\x25\xe9\x73\x5a\xa2\xcc\xd3\xc6\x7d\x14\x1f\x5a\x78\xb0\x0f\xbe\x91\xd7\x93\xb7\x7b\x09\x9c\x75\xa5\x99\x42\x2c\x8d\x48\x6c\x07\xfc\xf7\x18\xce\x37\x58\xa2\x18\xff\xd9\xec\x92\xcd\x74\x7d\xa8\x88\x07\x8e\xc7\x14\xd2\xa6\xe2\xb1\xa4\x38\xb4\x80\x15\xe7\xd6\x50\x2a\xcf\xbd\x74\xda\x06\x0f\xdc\x98\x41\x71\x80\xc9\xe9\xbf\x01\x00\x00\xff\xff\xb7\xb3\xd4\x2a\x71\x0b\x00\x00")

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

	info := bindataFileInfo{name: "swagger.json", size: 2929, mode: os.FileMode(420), modTime: time.Unix(1451994349, 0)}
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

