package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"forward/internal/socks5"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func init() {
	file, err := os.OpenFile("gen.yaml", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err == nil {
		mw := io.MultiWriter(os.Stdout, file)
		log.SetOutput(mw)
	}
	log.SetFlags(0)
}

func main() {
	resp, err := http.Get("https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt")
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		log.Println("err", err)
		return
	}

	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	body := string(all)
	decode, err := base64.StdEncoding.DecodeString(body)
	reader := bufio.NewReader(bytes.NewReader(decode))

	domainSuffix := make([]string, 0)
	for {
		line, _, _ := reader.ReadLine()
		if line == nil {
			break
		} else {
			s := string(line)
			suffix := getDomainSuffix(s)
			if suffix != "" {
				domainSuffix = append(domainSuffix, suffix)
			}
		}
	}

	rule := socks5.RouterRule{
		RuleType:     socks5.DOMAIN_SUFFIX,
		Values:       domainSuffix,
		UpstreamName: "bwg",
	}
	config := socks5.Config{
		Rules: []socks5.RouterRule{rule},
	}
	marshal, err := yaml.Marshal(config)
	log.Println(string(marshal))
}

func getDomainSuffix(s string) string {
	if strings.HasPrefix(s, "!") || strings.HasPrefix(s, "[") || len(s) == 0 || strings.HasPrefix(s, "@@") {
		return ""
	} else if strings.HasPrefix(s, "||") {
		domainSuffix := string([]byte(s)[2:])
		return domainSuffix
	} else if strings.HasPrefix(s, "|") {
		return ""
	} else {
		return s
	}
}
