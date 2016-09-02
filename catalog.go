package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

// CatalogFile contains catalog related meta-data
type CatalogFile struct {
	dataPath  string
	indexPath string
	file      *os.File
}

// open catalog file
func OpenCatalog(path string) (*CatalogFile, error) {
	data, index := splitCatalog(path)

	f, err := os.OpenFile(index, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	cf := new(CatalogFile)
	cf.indexPath = index
	cf.dataPath = data
	cf.file = f
	return cf, nil
}

// close catalog file
func (cf *CatalogFile) Close() error {
	if cf.file != nil {
		err := cf.file.Close()
		cf.file = nil
		return err
	}

	return nil // already closed
}

// add item to catalog atomically
func (cf *CatalogFile) AddFile(filename string, length uint64) (offset uint64, err error) {
	err = cf.lock()
	if err != nil {
		return
	}
	defer cf.unlock()

	type Header struct {
		Signature       uint32
		TotalItemsCount uint32
		TotalDataLength uint64
	}

	// TODO: check errors

	header := &Header{}
	order := binary.LittleEndian
	cf.file.Seek(0, 0) // begin
	err = binary.Read(cf.file, order, header)
	if err != nil {
		header.Signature = 0xafbeadde
	}

	// TODO: check signature for various versions
	offset = header.TotalDataLength
	header.TotalItemsCount += 1
	header.TotalDataLength += length

	cf.file.Seek(0, 0) // begin
	binary.Write(cf.file, order, header)

	cf.file.Seek(0, 2) // end
	cf.file.WriteString(fmt.Sprintf("%s,%d,%d,0\n", filename, offset, length))

	return
}

// get a few meta names from catalog name
func splitCatalog(catalog string) (data, index string) {
	// catalog = dir + file + ext
	dir, file := filepath.Split(catalog)
	// ext := filepath.Ext(file)
	// file = strings.TrimSuffix(file, ext)

	data = filepath.Join(dir, fmt.Sprintf("%s", file))
	index = filepath.Join(dir, fmt.Sprintf(".index-%s.txt", file))
	// lock = filepath.Join(dir, fmt.Sprintf(".lock%s", file))

	return
}

// try to acquire file lock
func (cf *CatalogFile) tryLock() (bool, error) {
	// assert(cf.file != nil)
	err := syscall.Flock(int(cf.file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	switch err {
	case nil:
		return true, nil
	case syscall.EWOULDBLOCK:
		return false, nil
	}

	return false, err
}

// acquire file lock
func (cf *CatalogFile) lock() error {
	// assert(cf.file != nil)
	return syscall.Flock(int(cf.file.Fd()), syscall.LOCK_EX)
}

// unlock file lock
func (cf *CatalogFile) unlock() error {
	// assert(cf.file != nil)
	return syscall.Flock(int(cf.file.Fd()), syscall.LOCK_UN)
}

// writes file to the catalog
func updateCatalog(mountPoint string, catalog, filename string, content io.Reader, length int64) (string, uint64, error) {
	if length < 0 {
		// save to temp file to determine data length
		tmp, err := ioutil.TempFile("", "temp_file")
		if err != nil {
			return "", 0, fmt.Errorf("failed to create temp file: %s", err)
		}
		defer func() {
			tmp.Close()
			os.RemoveAll(tmp.Name())
		}()

		length, err = io.Copy(tmp, content)
		if err != nil {
			return "", 0, fmt.Errorf("failed to copy content to temp file: %s", err)
		}
		tmp.Seek(0, 0) // go to begin
		content = tmp
	}

	// open and update catalog atomically
	cf, err := OpenCatalog(filepath.Join(mountPoint, catalog))
	if err != nil {
		return "", 0, err
	}
	offset, err := cf.AddFile(filename, uint64(length))
	if err != nil {
		return "", 0, err
	}

	dataPath := cf.dataPath

	// done index update
	data, err := os.OpenFile(dataPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return "", 0, fmt.Errorf("failed to open data file: %s", err)
	}
	defer data.Close()

	data.Seek(int64(offset), 0)
	n, err := io.Copy(data, content)
	if err != nil {
		return "", 0, fmt.Errorf("failed to copy data: %s", err)
	}
	if n != length {
		return "", 0, fmt.Errorf("only %d bytes copied of %d", n, length)
	}

	dataRel, _ := filepath.Rel(mountPoint, dataPath)
	return dataRel, uint64(length), nil // OK
}
