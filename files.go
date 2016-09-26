package main

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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/getryft/ryft-server/codec"
	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/catalog"
	"github.com/gin-gonic/gin"
)

// GetFileParams query parameters for GET /files
type GetFilesParams struct {
	Dir   string `form:"dir" json:"dir"`
	Local bool   `form:"local" json:"local"`
}

// DeleteFilesParams query parameters for DELETE /files
// there is no actual difference between dirs and files - everything will be deleted
type DeleteFilesParams struct {
	Files    []string `form:"file" json:"file"`
	Dirs     []string `form:"dir" json:"dir"`
	Catalogs []string `form:"catalog" json:"catalog"`
	Local    bool     `form:"local" json:"local"`
}

// to string
func (p DeleteFilesParams) String() string {
	return fmt.Sprintf("{files:%s, dirs:%s, catalogs:%s}",
		p.Files, p.Dirs, p.Catalogs)
}

// check is empty
func (p DeleteFilesParams) isEmpty() bool {
	return len(p.Files) == 0 &&
		len(p.Dirs) == 0 &&
		len(p.Catalogs) == 0
}

// NewFilesParams query parameters for POST /files
type NewFilesParams struct {
	Catalog   string `form:"catalog" json:"catalog"`     // catalog to save to
	Delimiter string `form:"delimiter" json:"delimiter"` // data delimiter
	File      string `form:"file" json:"file"`           // filename to save
	Offset    int64  `form:"offset" json:"offset"`       // offset inside file, used to rewrite
	Length    int64  `form:"length" json:"length"`       // data length
	Local     bool   `form:"local" json:"local"`
}

// GET /files method
func (s *Server) getFiles(ctx *gin.Context) {
	defer RecoverFromPanic(ctx)

	// parse request parameters
	params := GetFilesParams{}
	if err := ctx.Bind(&params); err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "failed to parse request parameters"))
	}

	// get search engine
	userName, authToken, homeDir, userTag := s.parseAuthAndHome(ctx)
	engine, err := s.getSearchEngine(params.Local, nil /*no files*/, authToken, homeDir, userTag)
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to get search engine"))
	}

	accept := ctx.NegotiateFormat(codec.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = codec.MIME_JSON
	}
	if accept != codec.MIME_JSON { //if accept == encoder.MIME_MSGPACK || accept == encoder.MIME_XMSGPACK {
		panic(NewServerError(http.StatusUnsupportedMediaType,
			"Only JSON format is supported for now"))
	}

	log.WithField("dir", params.Dir).WithField("user", userName).
		WithField("home", homeDir).WithField("cluster", userTag).
		Infof("start /files")
	info, err := engine.Files(params.Dir)
	if err != nil {
		// TODO: detail description?
		panic(NewServerError(http.StatusNotFound, err.Error()))
	}

	// TODO: if params.Sort {
	// sort names in the ascending order
	sort.Strings(info.Files)
	sort.Strings(info.Dirs)

	// TODO: use transcoder/dedicated structure instead of simple map!
	json := map[string]interface{}{
		"dir":     info.Path,
		"files":   info.Files,
		"folders": info.Dirs,
	}
	ctx.JSON(http.StatusOK, json)
}

// DELETE /files method
/* to test method:
curl -X DELETE -s "http://localhost:8765/files?file=p*.txt" | jq .
*/
func (s *Server) deleteFiles(ctx *gin.Context) {
	defer RecoverFromPanic(ctx)

	// parse request parameters
	params := DeleteFilesParams{}
	if err := ctx.Bind(&params); err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "failed to parse request parameters"))
	}

	userName, authToken, homeDir, userTag := s.parseAuthAndHome(ctx)
	mountPoint, err := s.getMountPoint(homeDir)
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to get mount point"))
	}

	log.WithField("dirs", params.Dirs).
		WithField("files", params.Files).
		WithField("catalogs", params.Catalogs).
		WithField("user", userName).
		WithField("home", homeDir).
		Info("deleting...")

	// for each requested file|dir|catalog get list of tags from consul KV/partition.
	// based of these tags determine the list of nodes having such file|dir|catalog.
	// for each node (with non empty list) call DELETE /files passing
	// list of files whose tags are matched.

	result := make(map[string]interface{})
	if !params.Local {
		files := params.Dirs[:]
		files = append(files, params.Files...)
		files = append(files, params.Catalogs...)

		services, tags, err := s.getConsulInfoForFiles(userTag, files)
		if err != nil || len(tags) != len(files) {
			panic(NewServerErrorWithDetails(http.StatusInternalServerError,
				err.Error(), "failed to map files to tags"))
		}

		type Node struct {
			IsLocal bool
			Name    string
			Address string
			Params  DeleteFilesParams

			Result interface{}
			Error  error
		}

		// build list of nodes to call
		nodes := make([]*Node, len(services))
		for k, f := range files {
			for i, service := range services {
				node := new(Node)
				scheme := "http"
				if port := service.ServicePort; port == 0 { // TODO: review the URL building!
					node.Address = fmt.Sprintf("%s://%s:8765", scheme, service.Address)
				} else {
					node.Address = fmt.Sprintf("%s://%s:%d", scheme, service.Address, port)
				}
				node.IsLocal = s.isLocalService(service)
				node.Name = service.Node
				node.Params.Local = true

				// check tags (no tags - all nodes)
				if len(tags[k]) == 0 || hasSomeTag(service.ServiceTags, tags[k]) {
					// based on 'k' index detect what the 'f' is: dir, file or catalog
					if k < len(params.Dirs) {
						node.Params.Dirs = append(node.Params.Dirs, f)
					} else if k < len(params.Dirs)+len(params.Files) {
						node.Params.Files = append(node.Params.Files, f)
					} else {
						node.Params.Catalogs = append(node.Params.Catalogs, f)
					}
				}

				nodes[i] = node
			}
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
					log.WithField("what", node.Params).Debugf("deleting on local node")
					node.Result, node.Error = s.deleteLocalFiles(mountPoint, node.Params), nil
				} else {
					log.WithField("what", node.Params).
						WithField("node", node.Name).
						WithField("addr", node.Address).
						Debugf("deleting on remote node")
					node.Result, node.Error = s.deleteRemoteFiles(node.Address, authToken, node.Params)
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
					"error": err.Error(),
				}
			} else {
				result[node.Name] = node.Result
			}
		}

	} else {
		result = s.deleteLocalFiles(mountPoint, params)
	}

	ctx.JSON(http.StatusOK, result)
}

// delete local nodes: files, dirs, catalogs
func (s *Server) deleteLocalFiles(mountPoint string, params DeleteFilesParams) map[string]interface{} {
	res := make(map[string]interface{})

	updateResult := func(name string, err error) {
		// in case of duplicate input
		// last result will be reported
		if err != nil {
			res[name] = err.Error()
		} else {
			res[name] = "OK" // "DELETED"
		}
	}

	// delete directories first ...
	for dir, err := range deleteAll(mountPoint, params.Dirs) {
		updateResult(dir, err)
	}

	// ... then delete files
	for file, err := range deleteAll(mountPoint, params.Files) {
		updateResult(file, err)
	}

	// ... then delete catalogs
	for cat, err := range deleteAllCatalogs(mountPoint, params.Catalogs) {
		updateResult(cat, err)
	}

	return res
}

// delete remote nodes: files, dirs, catalogs
func (s *Server) deleteRemoteFiles(address string, authToken string, params DeleteFilesParams) (map[string]interface{}, error) {
	// prepare query
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %s", err)
	}
	q := url.Values{}
	q.Set("local", fmt.Sprintf("%t", params.Local))
	for _, file := range params.Files {
		q.Add("file", file)
	}
	for _, dir := range params.Dirs {
		q.Add("dir", dir)
	}
	for _, catalog := range params.Catalogs {
		q.Add("catalog", catalog)
	}
	u.RawQuery = q.Encode()
	u.Path += "/files"

	// prepare request
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

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

// POST /files method
/* to test method:
curl -X POST -F file=@/path/to/file.txt -s "http://localhost:8765/files?file=/test/file\{\{random\}\}.txt" | jq .
curl -X POST --data "hello" -H 'Content-Type: application/octet-stream' -s "http://localhost:8765/files?file=/test/file\{\{random\}\}.txt" | jq .
*/
func (s *Server) postFiles(ctx *gin.Context) {
	defer RecoverFromPanic(ctx)

	// parse request parameters
	noDelim := fmt.Sprintf("no-binding-%x", time.Now().UnixNano()) // use random marker!
	params := NewFilesParams{}
	params.Delimiter = noDelim
	params.Offset = -1 // mark as "unspecified"
	params.Length = -1
	if err := ctx.Bind(&params); err != nil {
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			err.Error(), "failed to parse request parameters"))
	}

	// if delimiter is provided this value will be NOT NIL
	var delim *string
	if params.Delimiter != noDelim {
		delim = &params.Delimiter
	} else {
		params.Delimiter = ""
	}

	if len(params.File) == 0 {
		panic(NewServerError(http.StatusBadRequest,
			"no valid filename provided"))
	}

	userName, _, homeDir, _ := s.parseAuthAndHome(ctx)
	mountPoint, err := s.getMountPoint(homeDir)
	if err != nil {
		panic(NewServerErrorWithDetails(http.StatusInternalServerError,
			err.Error(), "failed to get mount point"))
	}

	var file io.Reader

	contentType := ctx.ContentType()
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "multipart/form-data":
		f, _, err := ctx.Request.FormFile("file")
		if err != nil {
			panic(NewServerErrorWithDetails(http.StatusBadRequest,
				err.Error(), `no "file" form data provided`))
		}
		defer f.Close()
		file = f
		log.Debugf("saving multipart form data...")

	case "application/octet-stream":
		file = ctx.Request.Body
		if params.Length < 0 { // if unspecified
			params.Length = ctx.Request.ContentLength
		}
		log.Debugf("saving octet-stream...")

	default:
		panic(NewServerErrorWithDetails(http.StatusBadRequest,
			contentType, "unexpected content type"))
	}

	response := map[string]interface{}{}
	log.WithField("params", params).
		WithField("user", userName).
		WithField("home", homeDir).
		Infof("saving new data...")
	status := http.StatusOK

	if len(params.Catalog) != 0 { // append to catalog
		catalog, length, err := updateCatalog(mountPoint, params, delim, file)

		if err != nil {
			status = http.StatusBadRequest // TODO: appropriate status code?
			response["error"] = err.Error()
			response["length"] = length
		} else {
			response["catalog"] = catalog
			response["length"] = length // not total, just this part
		}
	} else { // standalone file
		path, length, err := createFile(mountPoint, params, file)

		if err != nil {
			status = http.StatusBadRequest // TODO: appropriate status code?
			response["error"] = err.Error()
			response["length"] = length
		} else {
			response["path"] = path
			response["length"] = length
		}
	}

	ctx.JSON(status, response)
}

// get mount point path from local search engine
func (s *Server) getMountPoint(homeDir string) (string, error) {
	engine, err := s.getLocalSearchEngine(homeDir)
	if err != nil {
		return "", err
	}

	opts := engine.Options()
	return utils.AsString(opts["ryftone-mount"])
}

// remove directories or/and files
func deleteAll(mountPoint string, items []string) map[string]error {
	res := map[string]error{}
	for _, item := range items {
		path := filepath.Join(mountPoint, item)
		matches, err := filepath.Glob(path)
		if err != nil {
			res[item] = err
			continue
		}

		// remove all matches
		for _, file := range matches {
			rel, err := filepath.Rel(mountPoint, file)
			if err != nil {
				rel = file // ignore error and get absolute path
			}
			res[rel] = os.RemoveAll(file)
		}
	}

	return res
}

// remove catalogs
func deleteAllCatalogs(mountPoint string, items []string) map[string]error {
	res := map[string]error{}
	for _, item := range items {
		path := filepath.Join(mountPoint, item)
		matches, err := filepath.Glob(path)
		if err != nil {
			res[item] = err
			continue
		}

		// remove all matches
		for _, catalogPath := range matches {
			rel, err := filepath.Rel(mountPoint, catalogPath)
			if err != nil {
				rel = catalogPath // ignore error and get absolute path
			}

			// get catalog
			cat, err := catalog.OpenCatalog(catalogPath, true)
			if err != nil {
				res[rel] = err
				continue
			}
			defer func() {
				cat.DropFromCache()
				cat.Close() // it's ok to close later at function exit
				res[rel] = os.RemoveAll(catalogPath)
			}()

			// get data files
			files, err := cat.GetDataFiles()
			if err != nil {
				res[rel] = err
				continue
			}

			// make relative path
			for i, f := range files {
				if rf, err := filepath.Rel(mountPoint, f); err == nil {
					files[i] = rf
				}
			}

			// delete all data files
			for name, err := range deleteAll(mountPoint, files) {
				res[name] = err
			}
		}
	}

	return res
}

// createFile creates new file.
// Unique file name could be generated if path contains special keywords.
// Returns generated path (relative), length and error if any.
func createFile(mountPoint string, params NewFilesParams, content io.Reader) (string, uint64, error) {
	rbase := randomizePath(params.File) // first replace all {{random}} tokens
	rpath := rbase

	// create all parent directories
	pdir := filepath.Join(mountPoint, filepath.Dir(rpath))
	err := os.MkdirAll(pdir, 0755)
	if err != nil {
		return rpath, 0, fmt.Errorf("failed to create parent directories: %s", err)
	}

	var out *os.File
	flags := os.O_WRONLY | os.O_CREATE

	// if offset provided - file probably already exists
	// if no offset provided - file must not exist
	// if force flag is provided - we can override file
	if params.Offset < 0 /*&& !params.Force*/ {
		flags |= os.O_EXCL
	}

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
			return rpath, 0, err
		}

		break
	}

	defer out.Close()
	if 0 < params.Offset {
		_, err = out.Seek(params.Offset, 0)
		if err != nil {
			return rpath, 0, err
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
	return rpath, uint64(w), err
}

// append file to catalog
// Returns generated catalog path (relative), length and error if any.
func updateCatalog(mountPoint string, params NewFilesParams, delim *string, content io.Reader) (string, uint64, error) {
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
		tmp.Seek(0, 0) // go to begin
		content = tmp
	}

	// create all parent directories
	pdir := filepath.Join(mountPoint, filepath.Dir(catalogPath))
	err := os.MkdirAll(pdir, 0755)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create parent directories: %s", err)
	}

	// open catalog
	cat, err := catalog.OpenCatalog(filepath.Join(mountPoint, catalogPath), false)
	if err != nil {
		return "", 0, fmt.Errorf("failed to open catalog file: %s ", err)
	}
	defer cat.Close()

	// update catalog atomically
	data_path, data_pos, data_delim, err := cat.AddFile(filePath, params.Offset, params.Length, delim)
	if err != nil {
		return "", 0, fmt.Errorf("failed to add file to catalog: %s", err)
	}

	// TODO: in case of write error mark corresponding part as "bad"

	// write file content
	data, err := os.OpenFile(data_path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return "", 0, fmt.Errorf("failed to open data file: %s", err)
	}
	defer data.Close()

	_, err = data.Seek(int64(data_pos), 0)
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
