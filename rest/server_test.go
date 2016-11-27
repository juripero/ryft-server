package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	//"github.com/getryft/ryft-server/search"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gopkg.in/tylerb/graceful.v1"
)

var (
	testLogLevel = "debug"
	testFakePort = ":12345"
)

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

	if err := fs.server.Prepare(); err != nil {
		panic(err)
	}

	mux.GET("/search", fs.server.DoSearch)
	mux.GET("/count", fs.server.DoCount)
	mux.GET("/files", fs.server.DoGetFiles)

	return fs
}

// get request
func (fs *fakeServer) get(url string) ([]byte, int, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost%s%s", fs.worker.Addr, url), nil)
	if err != nil {
		return nil, 0, err // failed
	}

	// req.Header.Set("Accept", codec.MIME)
	//// authorization
	//if len(engine.AuthToken) != 0 {
	//	req.Header.Set("Authorization", engine.AuthToken)
	//}

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
