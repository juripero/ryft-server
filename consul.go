package main

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

//type Service struct {
//	Node           string   `json:"Node"`
//	Address        string   `json:"Address"`
//	ServiceID      string   `json:"ServiceID"`
//	ServiceName    string   `json:"ServiceName"`
//	ServiceAddress string   `json:"ServiceAddress"`
//	ServiceTags    []string `json:"ServiceTags"`
//	ServicePort    string   `json:"ServicePort"`
//}

func GetConsulInfo() (interface{}, error) {

	config := api.DefaultConfig()
	config.Datacenter = "dc1"
	client, err := api.NewClient(config)

	if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		catalog := client.Catalog()
		srvc, _, _ := catalog.Service("ryft-rest-api", "", nil)

		fmt.Println(srvc)
		return srvc, nil
	}
}
