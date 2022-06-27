package socks5

import (
	"bufio"
	"forward/internal/stream"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"net"
	url2 "net/url"
	"os"
	"strconv"
	"strings"
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

const DEFAULT = "default"
const FINAL = "final"

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
	defaultUpstream := config.UpstreamAlias[DEFAULT]
	log.Printf("default upstream is [%s]\n", defaultUpstream)
	if config.GfwText != "" && defaultUpstream != "" {
		rule := GenRouteRuleFromGfwText(config.GfwText, DEFAULT)
		if rule != nil {
			config.Rules = append(config.Rules, *rule)
		}
	}
	for _, upstream := range config.Upstreams {
		upstreamMap[upstream.Name] = &Upstream{Name: upstream.Name, Host: upstream.Host, Port: upstream.Port, BasicAuth: upstream.BasicAuth}
	}
	// 读取alias
	for alias := range config.UpstreamAlias {
		upstreamMap[alias] = upstreamMap[config.UpstreamAlias[alias]]
	}
	return nil
}

type HttpRequest struct {
	requestLine []byte
	method      string
	rawUrl      string
	urlParsed   *url2.URL
	version     string
	host        string
	port        int
	headers     [][]byte
	headerMap   map[string]string
}

var crlf = []byte("\r\n")

func handler(conn net.Conn) {
	reader := bufio.NewReader(conn)
	defer conn.Close()
	err := Handshake(reader, conn)
	_, isHttp := err.(StreamError)
	if err != nil && isHttp {
		// http处理
		line, isPrefix, err := reader.ReadLine()
		if isPrefix || err != nil {
			log.Println("error read http request line,", isPrefix, err)
			return
		}
		requestLine := string(line)
		split := strings.Split(requestLine, " ")
		if len(split) != 3 {
			return
		} else {
			request := &HttpRequest{
				requestLine: line,
				method:      split[0],
				rawUrl:      split[1],
				version:     split[2],
				headers:     make([][]byte, 0, 8),
				headerMap:   make(map[string]string),
			}
			if "CONNECT" == request.method {
				hostPort := strings.Split(request.rawUrl, ":")
				request.host = hostPort[0]
				if len(hostPort) == 2 {
					request.port, _ = strconv.Atoi(hostPort[1])
				} else {
					request.port = 443
				}
			} else {
				request.urlParsed, err = url2.Parse(request.rawUrl)
				if err != nil {
					return
				}
				if request.urlParsed.Port() == "" {
					request.port = 80
				} else {
					port, _ := strconv.Atoi(request.urlParsed.Port())
					request.port = port
				}
				request.host = request.urlParsed.Host
			}
			for {
				line, _, err := reader.ReadLine()
				if err != nil {
					return
				}
				if len(line) == 0 {
					break
				}
				headerKeyValue := strings.Split(string(line), ": ")
				if len(headerKeyValue) == 2 {
					request.headerMap[headerKeyValue[0]] = headerKeyValue[1]
				}
				request.headers = append(request.headers, line)
				if "CONNECT" != request.method {
					if headerKeyValue[0] == "Host" {
						hostAndPort := strings.Split(headerKeyValue[1], ":")
						request.host = hostAndPort[0]
						if len(hostAndPort) == 2 {
							request.port, err = strconv.Atoi(hostAndPort[1])
							if err != nil {
								request.port = 80
							}
						} else {
							request.port = 80
						}
					}
				}
			}
			upstream := determineUpstream(request.host)
			addr := request.host + ":" + strconv.Itoa(request.port)
			log.Printf("%21s => [%21s] => %s\n", conn.RemoteAddr().String(), InfoUpstream(upstream), addr)
			upstreamConn, err := buildOuterSocket(upstream, addr)
			if err != nil {
				return
			}
			if "CONNECT" == request.method {
				stream.WriteAll(conn, []byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
				stream.Relay(conn, upstreamConn, addr)
			} else if upstream != nil {
				stream.WriteAll(upstreamConn, request.requestLine)
				stream.WriteAll(upstreamConn, crlf)
				for _, header := range request.headers {
					stream.WriteAll(upstreamConn, header)
					stream.WriteAll(upstreamConn, crlf)
				}
				stream.WriteAll(upstreamConn, crlf)
				stream.Relay(conn, upstreamConn, addr)
			} else {
				stream.WriteAll(upstreamConn, []byte(request.method))
				stream.WriteAll(upstreamConn, []byte(" "))
				if request.urlParsed.RawQuery != "" {
					stream.WriteAll(upstreamConn, []byte(request.urlParsed.Path+"?"+request.urlParsed.RawQuery))
				} else {
					stream.WriteAll(upstreamConn, []byte(request.urlParsed.Path))
				}
				stream.WriteAll(upstreamConn, []byte(" "))
				stream.WriteAll(upstreamConn, []byte(request.version))
				stream.WriteAll(upstreamConn, crlf)
				delete(request.headerMap, "Proxy-Authorization")
				proxyConnection := request.headerMap["Proxy-Connection"]
				if proxyConnection != "" {
					request.headerMap["Connection"] = proxyConnection
				}
				for key, value := range request.headerMap {
					stream.WriteAll(upstreamConn, []byte(key))
					stream.WriteAll(upstreamConn, []byte(": "))
					stream.WriteAll(upstreamConn, []byte(value))
					stream.WriteAll(upstreamConn, crlf)
				}
				stream.WriteAll(upstreamConn, crlf)
				stream.Relay(conn, upstreamConn, addr)
			}
		}

		return
	} else {
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
	return upstreamMap[FINAL]
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
