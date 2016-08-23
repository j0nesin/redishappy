package types

import "errors"

type Consul struct {
	Address  string    `json:"address"`
	Scheme   string    `json:"scheme"`
	Cert     string    `json:"cert"`
	Key      string    `json:"key"`
	CACert   string    `json:"cacert"`
	Services []Service `json:"services"`
}

type Service struct {
	Cluster    string   `json:"cluster"`
	Node       string   `json:"node"`
	Tags       []string `json:"tags"`
	Datacenter string   `json:"datacenter"`
}

func (c Consul) FindByClusterName(name string) (Service, error) {
	for _, service := range c.Services {
		if service.Cluster == name {
			return service, nil
		}
	}
	return Service{}, errors.New("Service not found")
}
