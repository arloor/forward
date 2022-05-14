package stream

import (
	"crypto/tls"
	"io"
	"net"
	"strconv"
)

// Dial 建立到目标地址的TLS连接
func Dial(domain string, port int) (net.Conn, error) {
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
		io.Copy(conWithTarget, conWithClient)
		conWithTarget.Close()
		conWithClient.Close()
	}()
	io.Copy(conWithClient, conWithTarget)
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
