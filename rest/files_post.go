package rest

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/getryft/ryft-server/search/utils"
	"github.com/getryft/ryft-server/search/utils/catalog"
	"github.com/gin-gonic/gin"
)

// NewFilesParams query parameters for POST /files
type NewFilesParams struct {
	Catalog   string `form:"catalog" json:"catalog"`     // catalog to save to
	Delimiter string `form:"delimiter" json:"delimiter"` // data delimiter
	File      string `form:"file" json:"file"`           // filename to save
	Offset    int64  `form:"offset" json:"offset"`       // offset inside file, used to rewrite
	Length    int64  `form:"length" json:"length"`       // data length
	Local     bool   `form:"local" json:"local"`
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

	fw := getFileWriter(out.Name())
	defer fw.Release()

	// if offset provided - file probably already exists
	// if no offset provided - data will append!
	if params.Offset < 0 {
		if params.Length < 0 {
			return rpath, 0, fmt.Errorf("no valid length provided")
		}
		params.Offset, err = fw.Append(params.Length)
		if err != nil {
			return rpath, 0, err
		}
	}

	defer out.Close()
	if 0 <= params.Offset {
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
