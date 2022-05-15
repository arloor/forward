package socks5

import (
	"flag"
)

func init() {
	flag.StringVar(&socks5yaml, "socks5conf", "socks5.yaml", "socks5代理配置")
}

var (
	conf       Config
	socks5yaml string
)

type Config struct {
	LocalAddr string `yaml:"local-addr"`
	Upstream  string `yaml:"upstream"`
}
