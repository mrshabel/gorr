package main

import "log"

type UpstreamService struct {
	Name          string
	GatewayPrefix string
	URL           string
}

// upstream services
var services []UpstreamService = []UpstreamService{
	{Name: "service1", GatewayPrefix: "/service1", URL: "http://localhost:8002"},
	{Name: "service2", GatewayPrefix: "/service2", URL: "http://localhost:8003"},
}

func main() {
	proxy := Proxy{Address: "localhost:8000", Services: services}

	// proxy request to remote backend
	if err := proxy.Run(); err != nil {
		log.Fatal("failed to start reverse proxy server: ", err)
	}
}
