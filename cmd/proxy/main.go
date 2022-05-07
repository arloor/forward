package main

import (
	"github.com/caddyserver/caddy/caddy/caddymain"
	"io"
	"log"
	"os"

	_ "github.com/caddyserver/forwardproxy"
)

func init() {
	file := "/var/log/caddy.log"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
}

func main() {
	caddymain.EnableTelemetry = false
	caddymain.Run()
}
