package stream

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	pool = &sync.Pool{New: func() interface{} {
		return make([]byte, 32*1024)
	}}
	recvBytes   int64 = 0
	uploadBytes int64 = 0
)

// ErrShortWrite means that a write accepted fewer bytes than requested
// but failed to return an explicit error.
var ErrShortWrite = errors.New("short write")

// errInvalidWrite means that a write returned an impossible count.
var errInvalidWrite = errors.New("invalid write result")

// EOF is the error returned by Read when no more input is available.
// (Read must return EOF itself, not an error wrapping EOF,
// because callers will test for EOF using ==.)
// Functions should return EOF only to signal a graceful end of input.
// If the EOF occurs unexpectedly in a structured data stream,
// the appropriate error is either ErrUnexpectedEOF or some other error
// giving more detail.
var EOF = errors.New("EOF")

func PromMetrics(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(200)
	writer.Header().Add("Content-Type", "text/text; charset=utf-8")
	writer.Write([]byte("# HELP recv_bytes The total number of recv bytes\n# TYPE recv_bytes counter\nrecv_bytes "))
	writer.Write([]byte(strconv.FormatInt(atomic.LoadInt64(&recvBytes), 10)))
	writer.Write([]byte("\n"))
	writer.Write([]byte("# HELP upload_bytes The total number of upload bytes\n# TYPE upload_bytes counter\nupload_bytes "))
	writer.Write([]byte(strconv.FormatInt(atomic.LoadInt64(&uploadBytes), 10)))
	writer.Write([]byte("\n"))
}

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
		err := copyBufferWithCounter(writerOnly{conWithTarget}, conWithClient, buf, &uploadBytes)
		if err == io.ErrShortWrite {
			log.Println("copy short for", host)
		}
		conWithTarget.Close()
		conWithClient.Close()
	}()
	buf := pool.Get().([]byte)
	defer pool.Put(buf)
	err := copyBufferWithCounter(writerOnly{conWithClient}, conWithTarget, buf, &recvBytes)
	if err == io.ErrShortWrite {
		log.Println("copy short for", host)
	}
}
func copyBufferWithCounter(dst io.Writer, src io.Reader, buf []byte, counter *int64) (err error) {
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			atomic.AddInt64(counter, int64(nw))
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != EOF {
				err = er
			}
			break
		}
	}
	return err
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
