package auth

import (
	"encoding/json"
	"io/ioutil"

	"github.com/gin-gonic/gin"
)

func AuthBasicFile(fileName string) (gin.HandlerFunc, error) {
	users, err := checkPwd(fileName)
	if err != nil {
		return nil, err
	}
	return gin.BasicAuth(users), nil

}

func checkPwd(fileName string) (map[string]string, error) {
	var users map[string]string

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(data, &users)
	if err != nil {
		return nil, err
	}
	return users, nil
}
