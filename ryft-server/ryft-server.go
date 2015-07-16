package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/search-ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
	})

	r.GET("/search-fail", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusInternalServerError)
	})

	r.Run(":8765")
}

/* Help
https://golang.org/src/net/http/status.go -- statuses


*/
