package stream

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

var pool = &sync.Pool{New: func() interface{} {
	return make([]byte, 32*1024)
}}

var recvBytes = promauto.NewCounter(prometheus.CounterOpts{
	Name: "recv_bytes",
	Help: "The total number of processed bytes",
})

// 屏蔽掉TCPCon的ReadFrom接口，因为并不能使用SendFile来减少copy，反而不能使用io.CopyBuf的buf缓存了
type writerOnly struct {
	io.Writer
}

// dialTls 建立到目标地址的TLS连接
func dialTls(domain string, port int) (net.Conn, error) {
	tlsConf := tls.Config{
		ServerName: domain,
	}
	tlsConn, err := tls.Dial("tcp", domain+":"+strconv.Itoa(port), &tlsConf)
	if err != nil {
		return nil, err
	}
	err = tlsConn.Handshake()
	if err != nil {
		return nil, err
	}
	rawConn := tlsConn
	return rawConn, nil
}

func Relay(conWithClient, conWithTarget net.Conn, host string) {
	go func() {
		buf := pool.Get().([]byte)
		defer pool.Put(buf)
		_, err := io.CopyBuffer(writerOnly{conWithTarget}, conWithClient, buf)
		//recvBytes.Add(float64(n))
		if err == io.ErrShortWrite {
			log.Println("copy short for", host)
		}
		conWithTarget.Close()
		conWithClient.Close()
	}()
	buf := pool.Get().([]byte)
	defer pool.Put(buf)
	n, err := io.CopyBuffer(writerOnly{conWithClient}, conWithTarget, buf)
	recvBytes.Add(float64(n))
	if err == io.ErrShortWrite {
		log.Println("copy short for", host)
	}
}

func BuildUpstreamSocket(upstreamHost string, upstreamPort int, target string, basicAuth string) (conn net.Conn, err error) {
	conn, err = dialTls(upstreamHost, upstreamPort)
	if err != nil {
		return
	}
	err = handshakeWithUpstream(conn, target, basicAuth)
	if err != nil {
		return
	}
	return conn, nil
}

func handshakeWithUpstream(upstreamCon net.Conn, host string, basicAuth string) error {
	WriteAll(upstreamCon, []byte("CONNECT "+host+" HTTP/1.1\r\nHost: "+host+"\r\nProxy-Authorization: Basic "+basicAuth+"\r\n\r\n"))
	serverReader := bufio.NewReader(upstreamCon)
	var line []byte
	line, _, err := serverReader.ReadLine()
	statusLine := string(line)
	if err != nil || !strings.Contains(statusLine, "200") {
		return errors.New(fmt.Sprintf("upstream handshake error: %s %s", statusLine, err))
	}
	for {
		line, _, err = serverReader.ReadLine()
		if err != nil {
			return err
		}
		if len(line) == 0 {
			break
		}
	}
	return nil
}

func WriteAll(conn net.Conn, buf []byte) error {
	for writtenNum := 0; writtenNum != len(buf); {
		tempNum, err := conn.Write(buf[writtenNum:])
		if err != nil {
			return err
		}
		writtenNum += tempNum
	}
	return nil
}
