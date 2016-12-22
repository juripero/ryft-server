package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	//"github.com/getryft/ryft-server/search"
	"github.com/getryft/ryft-server/search/testfake"
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
		"home-dir":      "/ryft",
		"ryftone-mount": "/tmp",
		"host-name":     "",
	}
	fs.server.Config.SettingsPath = "/tmp/ryft/.settings"
	fs.server.Config.HostName = "node-1"

	fs.server.Config.ExtraRequest = true

	if err := fs.server.Prepare(); err != nil {
		panic(err)
	}

	mux.GET("/search", fs.server.DoSearch)
	mux.GET("/count", fs.server.DoCount)
	mux.GET("/files", fs.server.DoGetFiles)

	// DEBUG mode
	mux.GET("/logging/level", fs.server.DoLoggingLevel)

	return fs
}

// cleanup - delete whole home directory
func (fs *fakeServer) cleanup() {
	mount := fmt.Sprintf("%v", fs.server.Config.BackendOptions["ryftone-mount"])
	home := fmt.Sprintf("%v", fs.server.Config.BackendOptions["home-dir"])
	os.RemoveAll(filepath.Join(mount, home))
}

// get request
func (fs *fakeServer) get(url, accept string, cancelIn time.Duration) ([]byte, int, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost%s%s", fs.worker.Addr, url), nil)
	if err != nil {
		return nil, 0, err // failed
	}

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

// create engine
func TestServerCreate(t *testing.T) {
	server := NewServer() // valid (usual case)
	assert.NotNil(t, server)
}
