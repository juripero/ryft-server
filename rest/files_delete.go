package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/getryft/ryft-server/search/utils/catalog"

	"github.com/gin-gonic/gin"
)

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

// DELETE /files method
/* to test method:
curl -X DELETE -s "http://localhost:8765/files?file=p*.txt" | jq .
*/
func (s *Server) DoDeleteFiles(ctx *gin.Context) {
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
