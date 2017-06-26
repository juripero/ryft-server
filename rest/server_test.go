package rest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/getryft/ryft-server/search/testfake"
	"github.com/getryft/ryft-server/search/utils/catalog"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gopkg.in/tylerb/graceful.v1"
)

var (
	testLogLevel = "error"
	testFakePort = ":12345"
)

// set test log level
func testSetLogLevel() {
	// SetLogLevelString(testLogLevel)
	testfake.SetLogLevelString(testLogLevel)
}

// fake server to generate random data
type fakeServer struct {
	server *Server
	worker *graceful.Server
}

// create new fake server
func newFake() *fakeServer {
	gin.SetMode(gin.ReleaseMode)
	mux := gin.Default()

	fs := &fakeServer{
		server: NewServer(),
		worker: &graceful.Server{
			Timeout: 100 * time.Millisecond,
			Server: &http.Server{
				Addr:    testFakePort,
				Handler: mux,
			},
		},
	}

	// default configuration
	//fs.server.Config.BackendOptions
	//fs.server.Config.SearchBackend
	fs.server.Config.LocalOnly = true
	fs.server.Config.ListenAddress = testFakePort

	fs.server.Config.SearchBackend = testfake.TAG
	fs.server.Config.BackendOptions = map[string]interface{}{
		"instance-name": ".work",
		"home-dir":      "/",
		"ryftone-mount": "/tmp/ryft",
		"host-name":     "",
	}
	fs.server.Config.SettingsPath = "/tmp/ryft/.settings"
	fs.server.Config.HostName = "node-1"

	fs.server.Config.ExtraRequest = true

	if err := fs.server.Prepare(); err != nil {
		panic(err)
	}

	mux.GET("/search", fs.server.DoSearch)
	mux.GET("/search/show", fs.server.DoSearchShow)
	mux.GET("/count", fs.server.DoCount)
	mux.GET("/files", fs.server.DoGetFiles)
	mux.GET("/files/*path", fs.server.DoGetFiles)
	mux.POST("/files", fs.server.DoPostFiles)
	mux.POST("/files/*path", fs.server.DoPostFiles)
	mux.DELETE("/files", fs.server.DoDeleteFiles)
	mux.DELETE("/files/*path", fs.server.DoDeleteFiles)
	mux.PUT("/rename", fs.server.DoRenameFiles)
	mux.PUT("/rename/*path", fs.server.DoRenameFiles)

	mux.GET("/file", fs.server.DoGetFiles)         // alias used in integration tests
	mux.GET("/file/*path", fs.server.DoGetFiles)   // alias used in integration tests
	mux.POST("/file", fs.server.DoPostFiles)       // alias used in integration tests
	mux.POST("/file/*path", fs.server.DoPostFiles) // alias used in integration tests
	mux.POST("/raw", fs.server.DoPostFiles)        // alias used in integration tests
	mux.POST("/raw/*path", fs.server.DoPostFiles)  // alias used in integration tests

	// DEBUG mode
	mux.GET("/logging/level", fs.server.DoLoggingLevel)

	os.MkdirAll("/tmp/ryft/foo", 0755) // see BackendOptions above!
	ioutil.WriteFile("/tmp/ryft/1.txt", []byte(`
11111-hello-11111
22222-hello-22222
33333-hello-33333
44444-hello-44444
55555-hello-55555
`), 0644)

	ioutil.WriteFile("/tmp/ryft/foo/a.txt", []byte(`
11111-hello-11111
22222-hello-22222
33333-hello-33333
44444-hello-44444
55555-hello-55555
`), 0644)
	ioutil.WriteFile("/tmp/ryft/bad.dat", []byte(`hello`), 0222)

	if cat, err := catalog.OpenCatalogNoCache("/tmp/ryft/catalog.test"); err != nil {
		panic(err)
	} else {
		cat.DataSizeLimit = 50
		defer cat.Close()

		putData := func(filename string, data string) {
			dataPath, dataPos, delim, err := cat.AddFilePart(filename, -1, int64(len(data)), nil)
			if err != nil {
				panic(err)
			}

			dir, _ := filepath.Split(dataPath)
			os.MkdirAll(dir, 0755)
			f, err := os.OpenFile(dataPath, os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				panic(err)
			}

			defer f.Close()
			f.Seek(dataPos, os.SEEK_SET)
			f.Write([]byte(data))
			f.Write([]byte(delim))
		}

		// put 3 file parts to separate data files
		putData("1.txt", "11111-hello-11111")
		putData("2.txt", "22222-hello-22222")
		putData("3.txt", "33333-hello-33333")
		putData("1.txt", "aaaaa-hello-aaaaa")
		putData("2.txt", "bbbbb-hello-bbbbb")
		putData("3.txt", "ccccc-hello-ccccc")
	}

	return fs
}

// get home's directory
func (fs *fakeServer) homeDir() string {
	mount := fmt.Sprintf("%v", fs.server.Config.BackendOptions["ryftone-mount"])
	home := fmt.Sprintf("%v", fs.server.Config.BackendOptions["home-dir"])
	return filepath.Join(mount, home)
}

// cleanup - delete whole home directory
func (fs *fakeServer) cleanup() {
	os.RemoveAll(fs.homeDir())
	fs.server.Close()
}

// do a request
func (fs *fakeServer) do(req *http.Request, accept string, cancelIn time.Duration) ([]byte, int, error) {
	if len(accept) != 0 {
		req.Header.Set("Accept", accept)
	}
	//// authorization
	//if len(engine.AuthToken) != 0 {
	//	req.Header.Set("Authorization", engine.AuthToken)
	//}

	if cancelIn > 0 {
		ch := make(chan struct{})
		req.Cancel = ch
		go func() {
			time.Sleep(cancelIn)
			close(ch)
		}()
	}

	// do HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err // failed
	}

	defer resp.Body.Close() // close it later

	body, err := ioutil.ReadAll(resp.Body)
	return body, resp.StatusCode, err
}

// GET request
func (fs *fakeServer) GET(url, accept string, cancelIn time.Duration) ([]byte, int, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost%s%s", fs.worker.Addr, url), nil)
	if err != nil {
		return nil, 0, err // failed
	}

	return fs.do(req, accept, cancelIn)
}

// POST request
func (fs *fakeServer) POST(url, accept string, contentType, data string, cancelIn time.Duration) ([]byte, int, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost%s%s", fs.worker.Addr, url), bytes.NewBufferString(data))
	if err != nil {
		return nil, 0, err // failed
	}

	if len(contentType) != 0 {
		req.Header.Set("Content-Type", contentType)
	}

	return fs.do(req, accept, cancelIn)
}

// DELETE request
func (fs *fakeServer) DELETE(url, accept string, cancelIn time.Duration) ([]byte, int, error) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost%s%s", fs.worker.Addr, url), nil)
	if err != nil {
		return nil, 0, err // failed
	}

	return fs.do(req, accept, cancelIn)
}

// PUT request
func (fs *fakeServer) PUT(url, accept string, contentType, data string, cancelIn time.Duration) ([]byte, int, error) {
	req, err := http.NewRequest("PUT", fmt.Sprintf("http://localhost%s%s", fs.worker.Addr, url), bytes.NewBufferString(data))
	if err != nil {
		return nil, 0, err // failed
	}

	if len(contentType) != 0 {
		req.Header.Set("Content-Type", contentType)
	}

	return fs.do(req, accept, cancelIn)
}

// create engine
func TestServerCreate(t *testing.T) {
	server := NewServer() // valid (usual case)
	if assert.NotNil(t, server) {
		defer server.Close()
	}
}
