package models

type EnvoyConfig struct {
	Name string `json:"name"`
	Spec `json:"spec"`
}

type Spec struct {
	Listeners []Listener `json:"listeners"`
	Clusters  []Cluster  `json:"clusters"`
}

type Listener struct {
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Port    uint32  `json:"port"`
	Routes  []Route `json:"routes"`
}

type Route struct {
	Name         string   `json:"name"`
	Prefix       string   `json:"prefix"`
	ClusterNames []string `json:"clusters"`
}

type Cluster struct {
	Name      string     `json:"name"`
	Endpoints []Endpoint `json:"endpoints"`
}

type Endpoint struct {
	Address string `json:"address"`
	Port    uint32 `json:port"`
}
