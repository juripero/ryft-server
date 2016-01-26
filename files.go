package main

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
	Dir   string `form:"dir" json:"dir"`
	Local bool   `form:"local" json:"local"`
}

const (
	home        string = "/ryftone"
	filesName   string = "files"
	foldersName string = "folders"
)

func files(c *gin.Context) {
	defer srverr.Recover(c)

	var err error
	params := FilesParams{}
	if err = c.Bind(&params); err != nil {
		panic(srverr.New(http.StatusBadRequest, err.Error()))
	}

	dirPath := getDirPath(params.Dir)
	accept := c.NegotiateFormat(encoder.GetSupportedMimeTypes()...)
	// default to JSON
	if accept == "" {
		accept = encoder.MIMEJSON
	} else if accept == encoder.MIMEMSGPACK || accept == encoder.MIMEMSGPACKX {
		c.JSON(http.StatusUnsupportedMediaType, "Message pack not implemented yet")
		return
	}

	m, err := getNames(dirPath)
	if err != nil {
		panic(srverr.New(http.StatusNotFound, err.Error()))
	}

	c.JSON(http.StatusOK, m)

}

func getDirPath(dirPath string) string {
	if dirPath == "" {
		return home
	}
	return path.Join(home, dirPath)
}

func getNames(dirPath string) (map[string]interface{}, error) {

	var err error
	var items []os.FileInfo

	items, err = ioutil.ReadDir(dirPath)

	if err != nil {
		return nil, err
	}

	m := createNamesMap(items)
	if dirPath == home {
		m["dirPath"] = "/"
	} else {
		m["dirPath"] = strings.TrimPrefix(dirPath, home)

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
