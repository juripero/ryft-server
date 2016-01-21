package main

import "github.com/hashicorp/consul/api"

//type Service struct {
//	Node           string   `json:"Node"`
//	Address        string   `json:"Address"`
//	ServiceID      string   `json:"ServiceID"`
//	ServiceName    string   `json:"ServiceName"`
//	ServiceAddress string   `json:"ServiceAddress"`
//	ServiceTags    []string `json:"ServiceTags"`
//	ServicePort    string   `json:"ServicePort"`
//}

func GetConsulInfo() (address []*api.CatalogService, err error) {

	config := api.DefaultConfig()
	config.Datacenter = "dc1"
	client, err := api.NewClient(config)

	if err != nil {

		return nil, err
	}

	catalog := client.Catalog()
	srvc, _, _ := catalog.Service("ryft-rest-api", "", nil)

	// for _, value := range srvc {
	// 	address <- fmt.Sprintf("%v:%v", value.ServiceAddress, value.ServicePort)
	// }
	return srvc, err
}
