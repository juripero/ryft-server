/*
 * ============= Ryft-Customized BSD License ============
 * Copyright (c) 2015, Ryft Systems, Inc.
 * All rights reserved.
 * Redistribution and use in source and binary forms, with or without modification,
 * are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *   this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *   this list of conditions and the following disclaimer in the documentation and/or
 *   other materials provided with the distribution.
 * 3. All advertising materials mentioning features or use of this software must display the following acknowledgement:
 *   This product includes software developed by Ryft Systems, Inc.
 * 4. Neither the name of Ryft Systems, Inc. nor the names of its contributors may be used
 *   to endorse or promote products derived from this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY RYFT SYSTEMS, INC. ''AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL RYFT SYSTEMS, INC. BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 * ============
 */

package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/catalog"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// PostFilesParams query parameters for POST /files
type PostFilesParams struct {
	Catalog   string `form:"catalog" json:"catalog"`     // catalog to save to
	Delimiter string `form:"delimiter" json:"delimiter"` // data delimiter
	File      string `form:"file" json:"file"`           // filename to save
	Offset    int64  `form:"offset" json:"offset"`       // offset inside file, used to rewrite
	Length    int64  `form:"length" json:"length"`       // data length
	Local     bool   `form:"local" json:"local"`

	Lifetime string `form:"lifetime" json:"lifetime"` // optional file lifetime
	lifetime time.Duration

	ShareMode string `form:"share-mode" json:"share-mode"` // share mode to use
	shareMode utils.ShareMode
}

// is empty?
func (p PostFilesParams) isEmpty() bool {
	return len(p.Catalog) == 0 &&
		len(p.File) == 0
}

// to string
func (p PostFilesParams) String() string {
	res := make([]string, 0)

	// catalog
	if p.Catalog != "" {
		res = append(res, fmt.Sprintf("catalog:%s", p.Catalog))

		// delimiter
		if p.Delimiter != "" {
			res = append(res, fmt.Sprintf("delim:%s", p.Delimiter))
		}
	}

	// file
	if p.File != "" {
		res = append(res, fmt.Sprintf("file:%s", p.File))
	}

	// offset
	if p.Offset >= 0 {
		res = append(res, fmt.Sprintf("offset:%d", p.Offset))
	}

	// length
	if p.Length >= 0 {
		res = append(res, fmt.Sprintf("length:%d", p.Length))
	}

	// lifetime
	if p.Lifetime != "" {
		res = append(res, fmt.Sprintf("lifetime:%s", p.Lifetime))
	}

	if len(p.ShareMode) != 0 {
		res = append(res, fmt.Sprintf("share-mode:%s", p.ShareMode))
	}

	if p.Local {
		res = append(res, "local")
	}

	return fmt.Sprintf("{%s}", strings.Join(res, ", "))
}

// POST /files method
/* to test method:
curl -X POST -F file=@/path/to/file.txt -s "http://localhost:8765/files?file=/test/file\{\{random\}\}.txt" | jq .
curl -X POST --data "hello" -H 'Content-Type: application/octet-stream' -s "http://localhost:8765/files?file=/test/file\{\{random\}\}.txt" | jq .
*/
func (s *Server) DoPostFiles(ctx *gin.Context) {
	defer RecoverFromPanic(ctx)

	// parse request parameters
	noDelim := fmt.Sprintf("no-binding-%x", time.Now().UnixNano()) // use random marker!
	params := PostFilesParams{
		Delimiter: noDelim,
		Offset:    -1, // mark as "unspecified"
		Length:    -1,
	}
	b := binding.Default(ctx.Request.Method, ctx.ContentType())
	if err := b.Bind(ctx.Request, &params); err != nil {
		panic(NewError(http.StatusBadRequest, err.Error()).
			WithDetails("failed to parse request parameters"))
	}

	// if delimiter is provided this value will be NOT NIL
	var delim *string
	if params.Delimiter != noDelim {
		tmp := mustParseDelim(params.Delimiter)
		delim = &tmp
	} else {
		params.Delimiter = ""
		// delim is nil
	}

	if len(params.File) == 0 {
		panic(NewError(http.StatusBadRequest,
			"no valid filename provided"))
	}

	userName, authToken, homeDir, userTag := s.parseAuthAndHome(ctx)
	mountPoint, err := s.getMountPoint(homeDir)
	if err != nil {
		panic(NewError(http.StatusInternalServerError, err.Error()).
			WithDetails("failed to get mount point"))
	}
	mountPoint = filepath.Join(mountPoint, homeDir)

	// checks all the input filenames are relative to home
	if len(params.Catalog) != 0 {
		if !search.IsRelativeToHome(mountPoint, filepath.Join(mountPoint, params.Catalog)) {
			panic(NewError(http.StatusBadRequest,
				fmt.Sprintf("catalog path %q is not relative to home", params.Catalog)))
		}
	} else {
		if !search.IsRelativeToHome(mountPoint, filepath.Join(mountPoint, params.File)) {
			panic(NewError(http.StatusBadRequest,
				fmt.Sprintf("path %q is not relative to home", params.File)))
		}
	}

	var file io.Reader

	contentType := ctx.ContentType()
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "multipart/form-data":
		f, _, err := ctx.Request.FormFile("file")
		if err != nil {
			panic(NewError(http.StatusBadRequest, err.Error()).
				WithDetails(`no "file" form data provided`))
		}
		// Note, there is no automatic length
		defer f.Close()
		file = f

	case "application/octet-stream":
		file = ctx.Request.Body
		if params.Length < 0 { // if unspecified
			params.Length = ctx.Request.ContentLength
		}

	default:
		panic(NewError(http.StatusBadRequest, contentType).
			WithDetails("unexpected content type"))
	}

	if len(params.Lifetime) > 0 {
		if params.lifetime, err = time.ParseDuration(params.Lifetime); err != nil {
			panic(NewError(http.StatusBadRequest, err.Error()).
				WithDetails("failed to parse lifetime"))
		}
	}

	if len(params.ShareMode) > 0 {
		if params.shareMode, err = utils.SafeParseMode(params.ShareMode); err != nil {
			panic(NewError(http.StatusBadRequest, err.Error()).
				WithDetails("failed to parse share mode"))
		}
	}

	result := map[string]interface{}{}
	log.WithField("params", params).
		WithField("user", userName).
		WithField("home", homeDir).
		Infof("saving new %s data...", contentType)
	status := http.StatusOK

	if !params.Local && !s.Config.LocalOnly {
		files := []string{params.Catalog}
		if len(params.Catalog) == 0 {
			files[0] = params.File
		}

		services, tags, err := s.getConsulInfoForFiles(userTag, files)
		if err != nil || len(tags) != len(files) {
			panic(NewError(http.StatusInternalServerError,
				err.Error()).WithDetails("failed to map files to tags"))
		}
		log.WithField("tags", tags[0]).Debugf("related tags")

		type Node struct {
			IsLocal bool
			Name    string
			Address string

			Params PostFilesParams
			data   io.Reader

			Result map[string]interface{}
			Error  error
		}

		// build list of nodes to call
		nodes := make([]*Node, len(services))
		Ncopies := 0
		for i, service := range services {
			node := new(Node)
			scheme := "http"
			if port := service.ServicePort; port == 0 { // TODO: review the URL building!
				node.Address = fmt.Sprintf("%s://%s:8765", scheme, service.Address)
			} else {
				node.Address = fmt.Sprintf("%s://%s:%d", scheme, service.Address, port)
				// node.Name = fmt.Sprintf("%s-%d", service.Node, port)
			}
			node.IsLocal = s.isLocalService(service)
			node.Name = service.Node

			// check tags (no tags - all nodes)
			if len(tags[0]) == 0 || hasSomeTag(service.ServiceTags, tags[0]) {
				node.Params = params
				node.Params.Local = true
				Ncopies += 1
			}

			nodes[i] = node
		}

		minLen := []int64{}
		allPath := []string{}
		allCat := []string{}

		if Ncopies > 1 {
			// save to temp file to get multiple copies
			if len(catalog.DefaultTempDirectory) > 0 {
				_ = os.MkdirAll(catalog.DefaultTempDirectory, 0755)
			}
			tmp, err := ioutil.TempFile(catalog.DefaultTempDirectory, filepath.Base(params.File))
			if err != nil {
				panic(fmt.Errorf("failed to create temp file: %s", err))
			}
			defer func() {
				tmp.Close()
				os.RemoveAll(tmp.Name())
			}()

			var w int64
			if 0 < params.Length {
				w, err = io.CopyN(tmp, file, params.Length)
			} else {
				w, err = io.Copy(tmp, file)
			}
			if err != nil {
				panic(fmt.Errorf("failed to copy content to temp file: %s", err))
			}
			tmp.Seek(0, os.SEEK_SET /*TODO: io.SeekStart*/)

			// update node parameters
			for _, node := range nodes {
				if node.Params.isEmpty() {
					continue // nothing to do
				}

				f, err := os.Open(tmp.Name())
				if err != nil {
					panic(fmt.Errorf("failed to open temp file: %s", err))
				}
				defer f.Close()
				node.Params.Length = w // update length
				node.data = f
			}

			// call each node in dedicated goroutine
			var wg sync.WaitGroup
			for _, node := range nodes {
				if node.Params.isEmpty() {
					continue // nothing to do
				}

				wg.Add(1)
				go func(node *Node) {
					defer wg.Done()
					if node.IsLocal {
						log.WithField("what", node.Params).Debugf("copying on local node")
						_, node.Result, node.Error = s.postLocalFiles(mountPoint, node.Params, delim, node.data)
					} else {
						log.WithField("what", node.Params).
							WithField("node", node.Name).
							WithField("addr", node.Address).
							Debugf("copying on remote node")
						node.Result, node.Error = s.postRemoteFiles(node.Address, authToken, node.Params, delim, node.data)
					}
				}(node)
			}

			// wait and report all results
			wg.Wait()
			for _, node := range nodes {
				if node.Params.isEmpty() {
					continue // nothing to do
				}

				if node.Error != nil {
					result[node.Name] = map[string]interface{}{
						"error": node.Error.Error(),
					}
				} else {
					result[node.Name] = node.Result
				}

				// combine results
				if x, err := utils.AsUint64(node.Result["length"]); err == nil {
					minLen = append(minLen, int64(x))
				}
				if x, err := utils.AsString(node.Result["path"]); err == nil {
					allPath = append(allPath, x)
				}
				if x, err := utils.AsString(node.Result["catalog"]); err == nil {
					allCat = append(allCat, x)
				}
			}
		} else {
			for _, node := range nodes {
				if node.Params.isEmpty() {
					continue // nothing to do
				}

				if node.IsLocal {
					log.WithField("what", node.Params).Debugf("*copying on local node")
					_, node.Result, node.Error = s.postLocalFiles(mountPoint, node.Params, delim, file)
				} else {
					log.WithField("what", node.Params).
						WithField("node", node.Name).
						WithField("addr", node.Address).
						Debugf("*copying on remote node")
					node.Result, node.Error = s.postRemoteFiles(node.Address, authToken, node.Params, delim, file)
				}

				if node.Error != nil {
					result[node.Name] = map[string]interface{}{
						"error": err.Error(),
					}
				} else {
					result[node.Name] = node.Result
				}

				// combine results
				if x, err := utils.AsUint64(node.Result["length"]); err == nil {
					minLen = append(minLen, int64(x))
				}
				if x, err := utils.AsString(node.Result["path"]); err == nil {
					allPath = append(allPath, x)
				}
				if x, err := utils.AsString(node.Result["catalog"]); err == nil {
					allCat = append(allCat, x)
				}
				break // one node enough
			}
		}

		result = map[string]interface{}{
			"details": result,
			"length":  findMinLength(minLen),
		}
		if x := getUniqueOrEmpty(allPath); len(x) > 0 {
			result["path"] = x
		}
		if x := getUniqueOrEmpty(allCat); len(x) > 0 {
			result["catalog"] = x
		}
	} else {
		status, result, _ = s.postLocalFiles(mountPoint, params, delim, file)
	}

	ctx.JSON(status, result)
}

// post local nodes: files, dirs, catalogs
func (s *Server) postLocalFiles(mountPoint string, params PostFilesParams, delim *string, file io.Reader) (int, map[string]interface{}, error) {
	res := make(map[string]interface{})
	status := http.StatusOK

	if len(params.Catalog) != 0 { // append to catalog
		catalog, length, err := updateCatalog(mountPoint, params, delim, file)

		if err != nil {
			status = http.StatusBadRequest // TODO: appropriate status code?
			res["error"] = err.Error()
			res["length"] = length
		} else {
			if params.lifetime > 0 {
				s.addJob("delete-catalog",
					filepath.Join(mountPoint, catalog),
					time.Now().Add(params.lifetime))
			}
			res["catalog"] = catalog
			res["length"] = length // not total, just this part
		}

		return status, res, err
	} else { // standalone file
		path, offset, length, err := createFile(mountPoint, params, file)

		if err != nil {
			status = http.StatusBadRequest // TODO: appropriate status code?
			res["error"] = err.Error()
			res["length"] = length
			res["offset"] = offset
		} else {
			if params.lifetime > 0 {
				s.addJob("delete-file",
					filepath.Join(mountPoint, path),
					time.Now().Add(params.lifetime))
			}
			res["path"] = path
			res["length"] = length
			res["offset"] = offset
		}

		return status, res, err
	}
}

// post remote nodes: files
func (s *Server) postRemoteFiles(address string, authToken string, params PostFilesParams, delim *string, file io.Reader) (map[string]interface{}, error) {
	// prepare query
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %s", err)
	}
	q := url.Values{}
	q.Set("local", fmt.Sprintf("%t", params.Local))
	if len(params.Catalog) > 0 {
		q.Add("catalog", params.Catalog)
	}
	if delim != nil {
		q.Add("delimiter", *delim)
	}
	if len(params.File) > 0 {
		q.Add("file", params.File)
	}

	if 0 <= params.Offset {
		q.Add("offset", fmt.Sprintf("%d", params.Offset))
	}
	if 0 <= params.Length {
		q.Add("length", fmt.Sprintf("%d", params.Length))
	}
	if len(params.Lifetime) > 0 {
		q.Add("lifetime", params.Lifetime)
	}
	if len(params.ShareMode) > 0 {
		q.Add("share-mode", params.ShareMode)
	}
	u.RawQuery = q.Encode()
	u.Path += "/files"

	// prepare request
	req, err := http.NewRequest("POST", u.String(), file)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	// authorization
	if len(authToken) != 0 {
		req.Header.Set("Authorization", authToken)
	}

	// do HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %s", err)
	}

	defer resp.Body.Close() // close it later

	// check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid HTTP response status: %d (%s)", resp.StatusCode, resp.Status)
	}

	res := make(map[string]interface{})
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %s", err)
	}

	return res, nil // OK
}

// file lock system
type FileWriter struct {
	path   string
	refs   int32
	length int64
	lock   sync.Mutex
}

var (
	fileWriters     = make(map[string]*FileWriter)
	fileWritersLock sync.Mutex
)

// get file writer
func getFileWriter(path string) *FileWriter {
	fileWritersLock.Lock()
	defer fileWritersLock.Unlock()

	// find existing
	if fw, ok := fileWriters[path]; ok && fw != nil {
		fw.Acquire()
		return fw
	}

	// create new one
	fw := new(FileWriter)
	fw.path = path
	fw.length = -1 // unknown
	fileWriters[path] = fw
	fw.Acquire()
	return fw
}

// acquire reference
func (fw *FileWriter) Acquire() {
	atomic.AddInt32(&fw.refs, +1)
}

// release reference
func (fw *FileWriter) Release() {
	if atomic.AddInt32(&fw.refs, -1) == 0 {
		fileWritersLock.Lock()
		defer fileWritersLock.Unlock()

		// if no more references just delete
		delete(fileWriters, fw.path)
	}
}

// append a part, return offset
func (fw *FileWriter) Append(length int64) (int64, error) {
	if length < 0 {
		// TODO: lock the whole file until write is finished
		return 0, fmt.Errorf("unknown length")
	}

	fw.lock.Lock()
	defer fw.lock.Unlock()

	if fw.length < 0 {
		// length is unknown, update the length
		info, err := os.Stat(fw.path)
		if err != nil {
			if os.IsNotExist(err) {
				fw.length = 0
			} else {
				return 0, err
			}
		} else {
			fw.length = info.Size()
		}
	}

	// length is known
	offset := fw.length
	fw.length += length

	return offset, nil // OK
}

// createFile creates new file.
// Unique file name could be generated if path contains special keywords.
// Returns generated path (relative), length and error if any.
func createFile(mountPoint string, params PostFilesParams, content io.Reader) (string, int64, int64, error) {
	rbase := randomizePath(params.File) // first replace all {{random}} tokens
	rpath := rbase

	// create all parent directories
	pdir := filepath.Join(mountPoint, filepath.Dir(rpath))
	err := os.MkdirAll(pdir, 0755)
	if err != nil {
		return rpath, 0, 0, fmt.Errorf("failed to create parent directories: %s", err)
	}

	var out *os.File
	flags := os.O_WRONLY | os.O_CREATE

	// try to create file, if file already exists try with updated name
	for k := 0; ; k++ {
		fullpath := filepath.Join(mountPoint, rpath)
		out, err = os.OpenFile(fullpath, flags, 0644)
		if err != nil {
			if params.File != rbase && os.IsExist(err) {
				// generate new unique name
				ext := filepath.Ext(rbase)
				base := strings.TrimSuffix(rbase, ext)
				rpath = fmt.Sprintf("%s-%d%s", base, k+1, ext)

				continue
			}
			return rpath, 0, 0, err
		}

		break
	}
	defer out.Close()

	if !params.shareMode.IsIgnore() {
		// get "write" lock, fail if busy
		if utils.SafeLockWrite(out.Name(), params.shareMode) {
			defer utils.SafeUnlockWrite(out.Name())
		} else {
			return rpath, 0, 0, fmt.Errorf("%s file is busy", out.Name())
		}
	}

	fw := getFileWriter(out.Name())
	defer fw.Release()

	// if offset provided - file probably already exists
	// if no offset provided - data will append!
	if params.Offset < 0 {
		if params.Length < 0 {
			return rpath, 0, 0, fmt.Errorf("no valid length provided")
		}
		params.Offset, err = fw.Append(params.Length)
		if err != nil {
			return rpath, 0, 0, err
		}
	}

	if 0 <= params.Offset {
		_, err = out.Seek(params.Offset, os.SEEK_SET /*TODO: io.SeekStart*/)
		if err != nil {
			return rpath, 0, 0, err
		}
	}

	// copy the file content
	var w int64
	if 0 < params.Length {
		w, err = io.CopyN(out, content, params.Length)
	} else {
		w, err = io.Copy(out, content)
	}
	if err != nil {
		log.WithError(err).WithField("file", rpath).
			Warnf("failed to save data")

		// do not leave partially saved data?
		// _ = os.RemoveAll(fullpath)
	}

	// return path to file without mountpoint
	return rpath, params.Offset, w, err
}

// append file to catalog
// Returns generated catalog path (relative), length and error if any.
func updateCatalog(mountPoint string, params PostFilesParams, delim *string, content io.Reader) (string, uint64, error) {
	catalogPath := randomizePath(params.Catalog)
	filePath := randomizePath(params.File)

	if params.Length < 0 {
		// save to temp file to determine data length
		if len(catalog.DefaultTempDirectory) > 0 {
			_ = os.MkdirAll(catalog.DefaultTempDirectory, 0755)
		}
		tmp, err := ioutil.TempFile(catalog.DefaultTempDirectory, filepath.Base(params.File))
		if err != nil {
			return "", 0, fmt.Errorf("failed to create temp file: %s", err)
		}
		defer func() {
			tmp.Close()
			os.RemoveAll(tmp.Name())
		}()

		params.Length, err = io.Copy(tmp, content)
		if err != nil {
			return "", 0, fmt.Errorf("failed to copy content to temp file: %s", err)
		}
		tmp.Seek(0, os.SEEK_SET /*TODO: io.SeekStart*/)
		content = tmp
	}

	// create all parent directories
	pdir := filepath.Join(mountPoint, filepath.Dir(catalogPath))
	if err := os.MkdirAll(pdir, 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create parent directories: %s", err)
	}

	// open catalog
	cat, err := catalog.OpenCatalog(filepath.Join(mountPoint, catalogPath))
	if err != nil {
		return "", 0, fmt.Errorf("failed to open catalog file: %s ", err)
	}
	defer cat.Close()

	// update catalog atomically
	data_path, data_pos, data_delim, err := cat.AddFilePart(filePath, params.Offset, params.Length, delim)
	if err != nil {
		return "", 0, fmt.Errorf("failed to add file to catalog: %s", err)
	}

	log.WithField("data_path", data_path).
		WithField("data_pos", data_pos).
		Debugf("writing data part with delim:%x", data_delim)
	// TODO: in case of write error mark corresponding part as "bad"

	data_dir, _ := filepath.Split(data_path)
	if err := os.MkdirAll(data_dir, 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create parent directories: %s", err)
	}

	if !params.shareMode.IsIgnore() {
		// get "write" lock, fail if busy
		if utils.SafeLockWrite(data_path, params.shareMode) {
			defer utils.SafeUnlockWrite(data_path)
		} else {
			return "", 0, fmt.Errorf("%s file is busy", data_path)
		}
	}

	// write file content
	data, err := os.OpenFile(data_path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return "", 0, fmt.Errorf("failed to open data file: %s", err)
	}
	defer data.Close()

	_, err = data.Seek(data_pos, os.SEEK_SET /*TODO: io.SeekStart*/)
	if err != nil {
		return "", 0, fmt.Errorf("failed to seek data file: %s", err)
	}

	n, err := io.Copy(data, content)
	if err != nil {
		return "", 0, fmt.Errorf("failed to copy data: %s", err)
	}
	if n != params.Length {
		return "", 0, fmt.Errorf("only %d bytes copied of %d", n, params.Length)
	}

	// write data delimiter
	if len(data_delim) > 0 {
		nn, err := data.WriteString(data_delim)
		if err != nil {
			return "", 0, fmt.Errorf("failed to write delimiter: %s", err)
		}
		if nn != len(data_delim) {
			return "", 0, fmt.Errorf("only %d bytes copied of %d", nn, len(data_delim))
		}
	}

	// TODO: notify catalog write is done

	return catalogPath, uint64(n), nil // OK
}

// replace {{random}} sections of filename with random token.
// random token is based on current unix time in nanoseconds.
// multiple {{random}} are possible
func randomizePath(path string) string {
	token := func(string) string {
		return fmt.Sprintf("%016x", time.Now().UnixNano())
	}

	// TODO: use some hash here

	re := regexp.MustCompile(`{{random}}`)
	return re.ReplaceAllStringFunc(path, token)
}

// check intersection is non empty
func hasSomeTag(tags []string, what []string) bool {
	for _, t := range tags {
		for _, x := range what {
			if x == t {
				return true
			}
		}
	}

	return false
}

// find minimum length
func findMinLength(lens []int64) int64 {
	if len(lens) == 0 {
		return 0 // not found
	}

	res := lens[0]
	for i := 1; i < len(lens); i++ {
		if lens[i] < res {
			res = lens[i]
		}
	}

	return res
}

// find the unique string or empty
func getUniqueOrEmpty(all []string) string {
	unique := make(map[string]int)
	for _, s := range all {
		if len(s) > 0 {
			unique[s] = 1
		}
	}

	if len(unique) == 1 {
		for k, _ := range unique {
			return k // first key
		}
	}

	return "" // not unique
}
