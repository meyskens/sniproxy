package endpoints

type EndpointsAdvert struct {
	Endpoints []EndpointsAdvertEndpoint `json:"endpoints"`
}

type EndpointsAdvertEndpoint struct {
	Host   string `json:"host"`
	Remote string `json:"remote"`
}
