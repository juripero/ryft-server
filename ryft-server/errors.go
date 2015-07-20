package main

import (
	"fmt"
	"net/http"

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
			c.IndentedJSON(err.Status, gin.H{"message": err.Message, "status": err.Status})
			return
		}

		if err, ok := r.(error); ok {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error(), "status": http.StatusInternalServerError})
			return
		}

		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("%+v", r), "status": http.StatusInternalServerError})
	}
}
