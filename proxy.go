package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Proxy is the main reverse proxy server that accepts a list of upstream backend services
type Proxy struct {
	Address  string
	Services []UpstreamService
	Server   *http.Server
}

// Read/Write timeouts for proxy server
const RWTimeout = 10 * time.Second

// NewProxy creates and configures a new proxy
func NewProxy(address string, services []UpstreamService) *Proxy {
	return &Proxy{
		Address:  address,
		Services: services,
	}
}

// Run start the proxy server
func (p *Proxy) Run() error {
	// router to dynamically route handlers
	router := mux.NewRouter()

	// register service handlers
	for _, service := range p.Services {
		target, err := url.Parse(service.URL)
		if err != nil {
			return fmt.Errorf("invalid URL for service %s: %v", service.Name, err)
		}

		// create handler for the service
		handler := p.CreateHandler(service, target)

		// register the handler with its upstream gateway prefix

		router.PathPrefix(service.GatewayPrefix).HandlerFunc(handler)
		log.Printf("[Gorr] Register handler for upstream %s, prefix %s", service.Name, service.GatewayPrefix)
	}

	// create http server for proxy
	p.Server = &http.Server{
		Addr:         p.Address,
		Handler:      router,
		ReadTimeout:  RWTimeout,
		WriteTimeout: RWTimeout,
		IdleTimeout:  RWTimeout,
	}

	// listen to requests
	log.Println("[Gorr] proxy server started on: ", p.Address)
	return p.Server.ListenAndServe()
}

// CreateHandler creates a reverse proxy server and returns a new http handler for the given upstream service with its target URL. The incoming request details will be modified to ensure that the right headers and details are passed to the upstream server.
func (p *Proxy) CreateHandler(service UpstreamService, target *url.URL) func(w http.ResponseWriter, r *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(target)

	// setup director to modify the contents
	proxy.Director = func(inReq *http.Request) {
		originalReq := *inReq
		// update headers to match target
		inReq.Host = target.Host
		inReq.URL.Host = target.Host
		inReq.URL.Scheme = target.Scheme

		// strip gateway prefix from the target url path
		inReq.URL.Path = strings.TrimPrefix(inReq.URL.Path, service.GatewayPrefix)

		// clear any request uri
		inReq.RequestURI = ""
		// clear headers if any to prevent IP spoofing
		inReq.Header.Del("X-Forwarded-For")

		// log request
		log.Printf("[Gorr] Forwarding request: %s%s ->  %s%s\n", originalReq.Host, originalReq.URL.Path, inReq.Host, inReq.URL.Path)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// proxy request to upstream backend
		log.Printf("[Gorr] Received request: %s\n", r.URL.Path)
		proxy.ServeHTTP(w, r)
	}
}
