package socks5

import (
	"bufio"
	"encoding/base64"
	"forward/internal/stream"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var basicAuth string
var upstreamHost string
var upstreamPort int = 443

func Serve() {
	err := parseConf(socks5yaml)
	if err != nil {
		return
	}
	upstream, err := url.Parse(conf.Upstream)
	if conf.LocalAddr != "" && upstream.Host != "" && upstream.Scheme == "https" {
		parseUpstream(upstream)
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
	log.Println("\n" + string(bytes))
	err = yaml.Unmarshal(bytes, &conf)
	if err != nil {
		return err
	}
	return nil
}

func parseUpstream(upstream *url.URL) {
	username := upstream.User.Username()
	password, _ := upstream.User.Password()
	basicAuth = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	upstreamHost = upstream.Host
	if upstream.Port() == "" {
		upstreamPort = 443
	} else {
		upstreamPort, _ = strconv.Atoi(upstream.Port())
	}
}

func handler(conn net.Conn) {
	reader := bufio.NewReader(conn)
	defer conn.Close()
	handshakeErr := Handshake(reader, conn)
	if handshakeErr != nil {
		log.Println("handshakeErr ", handshakeErr)
		return
	}
	addr, port, getTargetErr := ParseRequest(reader, conn)
	if getTargetErr != nil {
		log.Println(getTargetErr)
		return
	}
	host := addr + ":" + strconv.Itoa(port)
	log.Println(conn.RemoteAddr().String(), "=>", host)
	serverConn, err := stream.Dial(upstreamHost, upstreamPort)
	if err != nil {
		log.Println(err)
		return
	}
	defer serverConn.Close()
	stream.WriteAll(serverConn, []byte("CONNECT "+host+" HTTP/1.1\r\nHost: "+host+"\r\nProxy-Authorization: Basic "+basicAuth+"\r\n\r\n"))
	serverReader := bufio.NewReader(serverConn)
	var line []byte
	line, _, err = serverReader.ReadLine()
	statusLine := string(line)
	if err != nil || !strings.Contains(statusLine, "200") {
		log.Println("与代理握手失败", statusLine, err)
		return
	}
	for {
		line, _, err = serverReader.ReadLine()
		if err != nil {
			return
		}
		if len(line) == 0 {
			break
		}
	}
	stream.Relay(conn, serverConn, host)
}

func listen(addr string) {
	if handler == nil {
		log.Println("handler为空，请先调用RegisterHandler")
		return
	}
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
