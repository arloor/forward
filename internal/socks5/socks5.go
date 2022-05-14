package socks5

import (
	"bufio"
	"encoding/base64"
	"forward/internal/stream"
	"log"
	"net"
	"strconv"
	"strings"
)

func Serve() {
	if proxyHost != "" {
		listen(":" + strconv.Itoa(socks5Port))
	}
}

func handler(conn net.Conn) {
	basicAuth := base64.StdEncoding.EncodeToString([]byte(proxyUser + ":" + proxyPasswd))
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
	log.Println("代理请求", host)
	serverConn, err := stream.Dial(proxyHost, proxyPort)
	if err != nil {
		log.Println(err)
		return
	}
	defer serverConn.Close()
	stream.WriteAll(serverConn, []byte("CONNECT "+host+" HTTP/1.1\r\nHost: "+host+"\r\nProxy-Authorization: Basic "+basicAuth+"\r\n\r\n"))
	serverReader := bufio.NewReader(serverConn)
	var line []byte
	line, _, err = serverReader.ReadLine()
	if err != nil || !strings.Contains(string(line), "200") {
		log.Println("与代理握手失败", err)
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
