// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package forwardproxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/caddyserver/forwardproxy/httpclient"
	"golang.org/x/net/proxy"
)

func Setup(host, port, upstream string) (*ForwardProxy, error) {
	fp := &ForwardProxy{
		dialTimeout: time.Second * 20,
		hostname:    host, port: port,
		httpTransport: http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConns:        50,
			IdleConnTimeout:     60 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
	fp.authCredentials = [][]byte{}
	fp.hideIP = true
	fp.hideVia = true
	fp.upstream, _ = url.Parse(upstream)

	dialer := &net.Dialer{
		Timeout:   fp.dialTimeout,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	fp.dialContext = dialer.DialContext
	fp.httpTransport.DialContext = func(ctx context.Context, network string, address string) (net.Conn, error) {
		conn, err := fp.dialContextCheckACL(ctx, network, address)
		if err != nil {
			return conn, err
		}
		return conn, nil
	}

	if fp.upstream != nil {
		if !isLocalhost(fp.upstream.Hostname()) && fp.upstream.Scheme != "https" {
			return nil, errors.New("insecure schemes are only allowed to localhost upstreams")
		}

		registerHTTPDialer := func(u *url.URL, _ proxy.Dialer) (proxy.Dialer, error) {
			// CONNECT request is proxied as-is, so we don't care about target url, but it could be
			// useful in future to implement policies of choosing between multiple upstream servers.
			// Given dialer is not used, since it's the same dialer provided by us.
			d, err := httpclient.NewHTTPConnectDialer(fp.upstream.String())
			if err != nil {
				return nil, err
			}
			d.Dialer = *dialer
			if isLocalhost(fp.upstream.Hostname()) && fp.upstream.Scheme == "https" {
				// disabling verification helps with testing the package and setups
				// either way, it's impossible to have a legit TLS certificate for "127.0.0.1"
				log.Println("Localhost upstream detected, disabling verification of TLS certificate")
				d.DialTLS = func(network string, address string) (net.Conn, string, error) {
					conn, err := tls.Dial(network, address, &tls.Config{InsecureSkipVerify: true})
					if err != nil {
						return nil, "", err
					}
					return conn, conn.ConnectionState().NegotiatedProtocol, nil
				}
			}
			return d, nil
		}
		proxy.RegisterDialerType("https", registerHTTPDialer)
		proxy.RegisterDialerType("http", registerHTTPDialer)

		upstreamDialer, err := proxy.FromURL(fp.upstream, dialer)
		if err != nil {
			return nil, errors.New("failed to create proxy to upstream: " + err.Error())
		}

		if ctxDialer, ok := upstreamDialer.(interface {
			DialContext(ctx context.Context, network, address string) (net.Conn, error)
		}); ok {
			// upstreamDialer has DialContext - use it
			fp.dialContext = ctxDialer.DialContext
		} else {
			// upstreamDialer does not have DialContext - ignore the context :(
			fp.dialContext = func(ctx context.Context, network string, address string) (net.Conn, error) {
				return upstreamDialer.Dial(network, address)
			}
		}
	}

	makeBuffer := func() interface{} { return make([]byte, 0, 32*1024) }
	bufferPool = sync.Pool{New: makeBuffer}
	return fp, nil
}

func isLocalhost(hostname string) bool {
	if hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" {
		return true
	}
	return false
}

func readLinesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var hostnames []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		hostnames = append(hostnames, scanner.Text())
	}

	return hostnames, scanner.Err()
}

// isValidDomainLite shamelessly rejects non-LDH names. returns nil if domains seems valid
func isValidDomainLite(domain string) error {
	for i := 0; i < len(domain); i++ {
		c := domain[i]
		if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_' || '0' <= c && c <= '9' ||
			c == '-' || c == '.' {
			continue
		}
		return errors.New("character " + string(c) + " is not allowed")
	}
	sections := strings.Split(domain, ".")
	for _, s := range sections {
		if len(s) == 0 {
			return errors.New("empty section between dots in domain name or trailing dot")
		}
		if len(s) > 63 {
			return errors.New("domain name section is too long")
		}
	}
	return nil
}

type ProxyError struct {
	S    string
	Code int
}

func (e *ProxyError) Error() string {
	return fmt.Sprintf("[%v] %s", e.Code, e.S)
}

func (e *ProxyError) SplitCodeError() (int, error) {
	if e == nil {
		return 200, nil
	}
	return e.Code, errors.New(e.S)
}

type Handler struct {
	Fp *ForwardProxy
}

func (ha *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ha.Fp.ServeHTTP(w, r)
}
