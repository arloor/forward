package socks5

import (
	"bufio"
	"forward/internal/stream"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net"
	"os"
	"strconv"
)

func Serve() {
	err := parseConf(socks5yaml)
	if err != nil {
		return
	}
	if config.LocalAddr != "" {
		listen(config.LocalAddr)
	}
}

func parseConf(socks5conf string) error {
	log.Println("read socks5 config from", socks5conf)
	file, err := os.Open(socks5conf)
	if err != nil {
		return err
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return err
	}
	if config.GfwText != "" && config.GfwUpstreamName != "" {
		rule := GenRouteRuleFromGfwText(config.GfwText, config.GfwUpstreamName)
		if rule != nil {
			config.Rules = append(config.Rules, *rule)
		}
	}
	for _, upstream := range config.Upstreams {
		upstreamMap[upstream.Name] = &Upstream{Name: upstream.Name, Host: upstream.Host, Port: upstream.Port, BasicAuth: upstream.BasicAuth}
	}
	return nil
}

func handler(conn net.Conn) {
	reader := bufio.NewReader(conn)
	defer conn.Close()
	err := Handshake(reader, conn)
	if err != nil {
		log.Println("socks5 handshake err: ", err)
		return
	}
	host, port, getTargetErr := ParseRequest(reader, conn)
	if getTargetErr != nil {
		log.Println("parse socks5 target err:", getTargetErr)
		return
	}
	addr := host + ":" + strconv.Itoa(port)
	upstream := determineUpstream(host)
	log.Printf("%21s => [%21s] => %s\n", conn.RemoteAddr().String(), InfoUpstream(upstream), addr)
	upstreamConn, err := buildOuterSocket(upstream, addr)
	if upstreamConn != nil {
		defer upstreamConn.Close()
	}
	if err != nil {
		log.Printf("%21s => [%21s] => %s error:%s\n", conn.RemoteAddr().String(), InfoUpstream(upstream), addr, err)
		return
	}
	stream.Relay(conn, upstreamConn, addr)
}

// 如果upstream为nil，则直连目标地址
func buildOuterSocket(upstream *Upstream, addr string) (conn net.Conn, err error) {
	if upstream != nil {
		return stream.BuildUpstreamSocket(upstream.Host, upstream.Port, addr, upstream.BasicAuth)
	} else {
		return net.Dial("tcp", addr)
	}
}

func determineUpstream(addr string) (upstream *Upstream) {
	ip := net.ParseIP(addr)
	for _, rule := range config.Rules {
		if rule.determine(addr, ip) {
			return upstreamMap[rule.UpstreamName]
		}
	}
	return upstreamMap[config.FinalUpstreamName]
}

func listen(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println("listen", addr, "error ", err)
		return
	}
	defer ln.Close()
	log.Println("serve socks5 proxy at ", ln.Addr())
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Println("accept socket ", err)
		} else {
			go handler(c)
		}
	}
}
