package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProxyRun(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T, proxy *Proxy)
	}{
		{"test request succeeds", testForwardRequest},
		{"test service not found", testServiceNotFound},
	}

	// mock upstream services
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			// reply with pong
			fmt.Fprintf(w, "pong")
			return
		}
		fmt.Fprint(w, "hello from mock server 1")
	}))

	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			// reply with pong
			fmt.Fprintf(w, "pong")
			return
		}
		fmt.Fprint(w, "hello from mocked server 2")
	}))

	services := []UpstreamService{
		{Name: "service1", GatewayPrefix: "/service1", URL: s1.URL},
		{Name: "service2", GatewayPrefix: "/service2", URL: s2.URL},
	}

	addr := "localhost:8000"
	// run the given test cases
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// create new proxy server
			proxy := NewProxy(addr, services)
			proxy.Run()
			test.fn(t, proxy)
			proxy.Server.Close()
		})
	}
}

func testForwardRequest(t *testing.T, proxy *Proxy) {
	// get random backend
	serverIdx := rand.Intn(len(proxy.Services) + 1)
	addr := fmt.Sprintf("http://%s/service%d/ping", proxy.Address, serverIdx)
	fmt.Println(addr)
	resp, err := http.Get(addr)
	if err != nil {
		t.Fatal("failed to reach upstream server", err)
	}
	defer resp.Body.Close()

	// check response status
	if resp.StatusCode != 200 {
		t.Fatal("received an expected response:", resp.Status)
	}

	// check response body

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("failed to read response body", err)
	}

	// assert that the response is "pong"
	assertEqual(t, string(body), "pong")
}

func testServiceNotFound(t *testing.T, proxy *Proxy) {
	addr := fmt.Sprintf("http://%s/service%d/ping", proxy.Address, len(proxy.Services)+1)
	resp, err := http.Get(addr)
	if err != nil {
		t.Fatal("failed to reach upstream server", err)
	}
	defer resp.Body.Close()

	// check response status
	if resp.StatusCode != 404 {
		t.Fatal("expected status 404, got:", resp.Status)
	}
}

func assertEqual(t *testing.T, want, got interface{}) {
	t.Helper()

	if want != got {
		t.Fatalf("wanted %v, but got %v", want, got)
	}
}
