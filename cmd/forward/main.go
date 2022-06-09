package main

import (
	"flag"
	"forward/internal/socks5"
	"forward/internal/stream"
	httpproxy "github.com/caddyserver/caddy/caddy/caddymain"
	_ "github.com/caddyserver/forwardproxy"
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
		http.HandleFunc("/metrics", stream.PromMetrics)
		http.HandleFunc("/final", socks5.ModifyFinalUpstream)
		http.ListenAndServe(":9999", nil)
	}()
	go socks5.Serve()
	httpproxy.EnableTelemetry = false
	httpproxy.Run()
}
