package main

import (
	"flag"
	"forward/internal/socks5"
	http "github.com/caddyserver/caddy/caddy/caddymain"
	_ "github.com/caddyserver/forwardproxy"
	"log"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Flags())
}

func main() {
	flag.Parse()
	go socks5.Serve()
	http.EnableTelemetry = false
	http.Run()
}
