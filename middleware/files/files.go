package files

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/getryft/ryft-server/encoder"
	"github.com/getryft/ryft-server/srverr"
	"github.com/gin-gonic/gin"
)

type FilesParams struct {
	Path  string `form:"path" json:"path"`
	Local bool   `form:"local" json:"local"`
}

const (
	home        string = "/ryftone"
	filesName   string = "files"
	foldersName string = "folders"
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

	m, err := getNames(path)
	if err != nil {
		panic(srverr.New(http.StatusNotFound, err.Error()))
	}

	c.JSON(http.StatusOK, m)

}

func getPath(folderPath string) string {
	if folderPath == "" {
		return home
	}
	return path.Join(home, folderPath)
}

func getNames(folderPath string) (map[string]interface{}, error) {

	var err error
	var items []os.FileInfo

	items, err = ioutil.ReadDir(folderPath)

	if err != nil {
		return nil, err
	}

	m := createNamesMap(items)
	if folderPath == home {
		m["path"] = "/"
	} else {
		m["path"] = strings.TrimPrefix(folderPath, home)

	}

	return m, nil
}

func createNamesMap(items []os.FileInfo) map[string]interface{} {
	m := map[string]interface{}{}
	files := []string{}
	folders := []string{}

	for _, v := range items {
		if strings.HasPrefix(v.Name(), ".") {
			continue
		}
		if v.IsDir() {
			folders = append(folders, v.Name())
		} else {
			files = append(files, v.Name())
		}
	}
	m[filesName] = files
	m[foldersName] = folders
	return m
}
