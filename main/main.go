package main

import (
	"fmt"
	"net/http"
)

const (
	// The port on which the SFC controller listens for HTTP traffic.
	port = "8100"
)

func main() {
	fmt.Printf("SFC-controller v0.0.2 Listening...\n")

	// Start server
	svr := &http.Server{
		Addr: ":" + port,
	}
	svr.Handler = http.HandlerFunc(handler)
	_ = svr.ListenAndServe()

	// Create Channel
	ch := make(chan bool)
	<-ch
}
