package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "requests",
		Help: "The total number of requests processed",
	})
)

func main() {

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":8080", nil)

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = false
	proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {

			// Check HTTP AUTH
			auth := req.Header.Get("Proxy-Authorization")
			if auth == "" {
				log.Println("ProxyAuthHeader not set")
				return req, goproxy.NewResponse(req,
					goproxy.ContentTypeText, http.StatusForbidden,
					"ProxyAuthHeader not set!")
			}
			const prefix = "Basic "
			if !strings.HasPrefix(auth, prefix) {
				return req, goproxy.NewResponse(req,
					goproxy.ContentTypeText, http.StatusForbidden,
					"ProxyAuthHeader not set!")
			}
			c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
			if err != nil {
				return req, goproxy.NewResponse(req,
					goproxy.ContentTypeText, http.StatusForbidden,
					"ProxyAuthHeader not set!")
			}
			cs := string(c)
			s := strings.IndexByte(cs, ':')
			if s < 0 {
				return req, goproxy.NewResponse(req,
					goproxy.ContentTypeText, http.StatusForbidden,
					"ProxyAuthHeader not set!")
			}

			// username , password
			//log.Println(cs[:s], cs[s+1:])

			// Setup Proxy Connection
			proxyUrl := fmt.Sprintf("http://%s:%s", cs[:s], cs[s+1:])

			//log.Println("tr... dial", proxyUrl)

			proxy.Tr = &http.Transport{Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse(proxyUrl)
			}}
			log.Println(req.URL)

			proxy.ConnectDial = proxy.NewConnectDialToProxy(proxyUrl)

			requestsProcessed.Inc()

			return req, nil
		})

	log.Println("serving end proxy server at 0.0.0.0:8001")
	log.Fatal(http.ListenAndServe(":8001", proxy))
}
