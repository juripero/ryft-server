package files

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/getryft/ryft-server/encoder"
	"github.com/getryft/ryft-server/srverr"
	"github.com/gin-gonic/gin"
)

type FilesParams struct {
	Path      string `form:"path" json:"path" `
	Local     bool   `form:"local" json:"local"`
	FilesOnly bool   `form:"fo" json:"fo"`
}

const (
	home      string = "/ryftone"
	arrayName string = "names"
	slash            = "/"
)

func Files(c *gin.Context) {
	defer srverr.DeferRecover(c)

	var err error
	params := FilesParams{}
	if err = c.Bind(&params); err != nil {
		panic(srverr.New(http.StatusBadRequest, err.Error()))
	}

	path := getPath(params.Path)
	accept := c.NegotiateFormat(encoder.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = encoder.MIMEJSON
	} else if accept == encoder.MIMEMSGPACK || accept == encoder.MIMEMSGPACKX {
		c.JSON(http.StatusUnsupportedMediaType, "Message pack not implemented yet")
		return
	}
	// var enc encoder.Encoder
	// if enc, err = encoder.GetByMimeType(accept); err != nil {
	// 	panic(srverr.New(http.StatusBadRequest, err.Error()))
	// }

	m, err := getNames(path, params.FilesOnly)
	if err != nil {
		c.JSON(http.StatusBadRequest, path)
		return
	}

	c.JSON(http.StatusOK, m)

}

func getPath(path string) string {
	if path == "" {
		return home
	}
	return home + slash + path
}

func getNames(path string, isAll bool) (map[string]interface{}, error) {

	var err error
	var items []os.FileInfo

	items, err = ioutil.ReadDir(path)

	if err != nil {
		return nil, err
	}

	if isAll {
		return getFilesOnly(items), nil
	}

	return getAll(items), nil
}

func getAll(items []os.FileInfo) map[string]interface{} {
	m := map[string]interface{}{}
	var names []string

	for _, v := range items {
		if strings.HasPrefix(v.Name(), ".") {
			continue
		}
		names = append(names, v.Name())
	}
	m[arrayName] = names
	return m
}

func getFilesOnly(items []os.FileInfo) map[string]interface{} {
	m := map[string]interface{}{}
	var names []string
	for _, v := range items {
		if strings.HasPrefix(v.Name(), ".") {
			continue
		}
		if !v.IsDir() {
			names = append(names, v.Name())
		}
	}
	m[arrayName] = names
	return m
}
