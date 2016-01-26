package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
)

func Consul(c *gin.Context) {

	config := api.DefaultConfig()

	client, err := api.NewClient(config)
	if err != nil {
		catalog := client.Catalog()
		node, _, _ := catalog.Node("ryftone-vm-selaptop-1", nil)
		nodesString := fmt.Sprintf("%v", node.Services)
		fmt.Println(nodesString)
		c.JSON(http.StatusOK, nodesString)
		return
	}
	c.JSON(http.StatusInternalServerError, "CONSUL FAIL")

}
