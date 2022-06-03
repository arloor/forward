package main

import (
	"flag"
	"forward/internal/socks5"
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
	go http.ListenAndServe("localhost:9999", nil)
	go socks5.Serve()
	httpproxy.EnableTelemetry = false
	httpproxy.Run()
}
