package main

import (
	"forward/internal/socks5"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"os"
)

func init() {
	file, err := os.OpenFile("gen.yaml", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err == nil {
		mw := io.MultiWriter(os.Stdout, file)
		log.SetOutput(mw)
	}
	log.SetFlags(0)
}

func main() {
	filename := "gfwlist.txt"
	socks5.DownloadGfwRaw(filename)
	rule := socks5.GenRouteRuleFromGfwText(filename, "gfw")
	config := socks5.Config{
		Rules: []socks5.RouterRule{*rule},
	}
	marshal, _ := yaml.Marshal(config)
	log.Println(string(marshal))
}
