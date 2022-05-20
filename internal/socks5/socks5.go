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
	if conf.LocalAddr != "" {
		listen(conf.LocalAddr)
	}
}

func parseConf(socks5conf string) error {
	log.Println("从", socks5conf, "读取socks5配置")
	file, err := os.Open(socks5conf)
	if err != nil {
		return err
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	log.Println("socks5启动配置为\n" + string(bytes))
	err = yaml.Unmarshal(bytes, &conf)
	if err != nil {
		return err
	}
	for _, upstream := range conf.Upstreams {
		upstreamMap[upstream.Name] = &Upstream{Name: upstream.Name, Host: upstream.Host, Port: upstream.Port, BasicAuth: upstream.BasicAuth}
	}
	return nil
}

func handler(conn net.Conn) {
	reader := bufio.NewReader(conn)
	defer conn.Close()
	err := Handshake(reader, conn)
	if err != nil {
		log.Println("err ", err)
		return
	}
	host, port, getTargetErr := ParseRequest(reader, conn)
	if getTargetErr != nil {
		log.Println(getTargetErr)
		return
	}
	addr := host + ":" + strconv.Itoa(port)
	upstream := determineUpstream(host)
	log.Println(conn.RemoteAddr().String(), "=> [", InfoUpstream(upstream), "] =>", addr)
	upstreamConn, err := buildOuterSocket(upstream, addr)
	if upstreamConn != nil {
		defer upstreamConn.Close()
	}
	if err != nil {
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
	for _, rule := range conf.Rules {
		if rule.determine(addr, ip) {
			return upstreamMap[rule.UpstreamName]
		}
	}
	return nil
}

func listen(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Println("监听", addr, "失败 ", err)
		return
	}
	defer ln.Close()
	log.Println("在 ", ln.Addr(), "启动socks代理")
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Println("接受连接失败 ", err)
		} else {
			go handler(c)
		}
	}
}
