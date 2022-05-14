package socks5

import (
	"flag"
)

func init() {
	flag.IntVar(&socks5Port, "socks5port", 1080, "启动socks5代理的端口")
	flag.StringVar(&proxyHost, "proxyHost", "", "代理地址")
	flag.IntVar(&proxyPort, "proxyPort", 443, "代理端口")
	flag.StringVar(&proxyUser, "proxyUser", "", "代理用户")
	flag.StringVar(&proxyPasswd, "proxyPasswd", "", "代理密码")
}

var (
	socks5Port  int
	proxyHost   string
	proxyPort   int
	proxyUser   string
	proxyPasswd string
)
