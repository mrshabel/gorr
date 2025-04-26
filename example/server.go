// this module contains a dummy server for trying out the reverse proxy
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	var port int64 = 8002
	// get the optional port
	if len(os.Args) > 1 {
		portStr := os.Args[1]
		var err error
		if port, err = strconv.ParseInt(portStr, 10, 0); err != nil {
			log.Fatal("invalid port number:", err)
		}
	}
	log.Println("example server listening on port", port)

	// handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("New request received from host=%s, address=%s", r.Host, r.RemoteAddr)

		// log headers and respond with pong
		log.Println("headers:", r.Header)
		log.Println("path:", r.URL.Path)

		if r.URL.Path == "/ping" {
			w.Write([]byte("pong"))
			return
		}

		w.Write([]byte("hello from example server"))
	})

	addr := fmt.Sprintf("localhost:%d", port)
	log.Fatal(http.ListenAndServe(addr, nil))
}
