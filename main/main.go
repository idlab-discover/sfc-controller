package main

import (
	"fmt"
	"net/http"
)

const (
	// port is the port on which the scheduler listens for HTTP traffic.
	port = "8100"
)

func main() {
	fmt.Printf("SFC-controller v0.0.1 Listening...\n")

	//start server
	svr := &http.Server{
		Addr: ":" + port,
	}
	svr.Handler = http.HandlerFunc(handler)
	svr.ListenAndServe()

	//make sure we live forever.
	ch := make(chan bool)
	<-ch
}
