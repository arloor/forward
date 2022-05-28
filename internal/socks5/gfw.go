package socks5

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func DownloadGfwRaw(filename string) {
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
		return
	}
	err = ioutil.WriteFile(filename, all, 0600)
}

func GenRouteRuleFromGfwText(filename string, gfwUpstreamName string) *RouterRule {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0666)
	if err != nil {
		return nil
	}
	defer file.Close()
	all, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(err)
		return nil
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

	rule := RouterRule{
		RuleType:     DOMAIN_SUFFIX,
		Values:       domainSuffix,
		UpstreamName: gfwUpstreamName,
	}
	return &rule
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
