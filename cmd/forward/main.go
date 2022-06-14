package main

import (
	"flag"
	forwardproxy "forward/internal/httpproxy"
	"forward/internal/socks5"
	"forward/internal/stream"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"net/http"

	_ "net/http/pprof"
	"os"
)

var (
	logFile  string
	httpAddr string
	upstream string
)

func init() {
	flag.StringVar(&logFile, "log", "stdout", "日志文件")
	flag.StringVar(&httpAddr, "http", "localhost:3128", "http代理监听地址")
	flag.StringVar(&upstream, "upstream", "socks5://localhost:1080", "http代理上游url")
}

func main() {
	flag.Parse()
	setupLog()
	go func() {
		http.HandleFunc("/metrics", stream.PromMetrics)
		http.HandleFunc("/final", socks5.ModifyFinalUpstream)
		http.HandleFunc("/", stream.ServeLine)
		http.ListenAndServe(":9999", nil)
	}()
	go socks5.Serve()
	forwardProxy, _ := forwardproxy.Setup("localhost", "3128", upstream)
	http.ListenAndServe(httpAddr, &forwardproxy.Handler{Fp: forwardProxy})
}

func setupLog() {
	if logFile == "stdout" {
		log.SetOutput(os.Stdout)
	} else if logFile == "stderr" {
		log.SetOutput(os.Stderr)
	} else {
		rollingFile := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    50,
			MaxAge:     14,
			MaxBackups: 10,
			Compress:   false,
		}
		mw := io.MultiWriter(os.Stdout, rollingFile)
		log.SetOutput(mw)
	}
	log.SetFlags(log.Lshortfile | log.Flags())
}
