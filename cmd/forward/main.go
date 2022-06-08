package main

import (
	"flag"
	"forward/internal/socks5"
	httpproxy "github.com/caddyserver/caddy/caddy/caddymain"
	_ "github.com/caddyserver/forwardproxy"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Flags())
}

func main() {
	flag.Parse()
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.HandleFunc("/final", socks5.ModifyFinalUpstream)
		http.ListenAndServe("localhost:9999", nil)
	}()
	go socks5.Serve()
	httpproxy.EnableTelemetry = false
	httpproxy.Run()
}
