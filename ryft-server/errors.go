// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

type ServerError struct {
	Status  int
	Message string
}

func (err *ServerError) Error() string {
	return fmt.Sprintf("%d %s", err.Status, err.Message)
}

func deferRecover(c *gin.Context) {
	if r := recover(); r != nil {
		if err, ok := r.(*ServerError); ok {
			log.Printf("Panic recovered server error: status=%d msg:%s", err.Status, err.Message)
			c.IndentedJSON(err.Status, gin.H{"message": err.Message, "status": err.Status})
			return
		}

		if err, ok := r.(error); ok {
			log.Printf("Panic recovered unknown error with msg:%s", err.Error())
			debug.PrintStack()
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error(), "status": http.StatusInternalServerError})
			return
		}

		log.Printf("Panic recovered with object:%+v", r)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("%+v", r), "status": http.StatusInternalServerError})
	}
}
