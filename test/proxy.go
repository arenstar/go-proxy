package main

import (
	"log"
	"net/http"

	"github.com/elazarl/goproxy"
)

func main() {

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false
	log.Println("serving end proxy server at 0.0.0.0:8002")
	log.Fatal(http.ListenAndServe(":8002", proxy))
}
