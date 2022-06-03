package socks5

import (
	"flag"
	"gopkg.in/yaml.v2"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

const DOMAIN_SUFFIX = "DOMAIN-SUFFIX"
const IP_CIDR = "IP-CIDR"
const MATCH = "MATCH"

func init() {
	flag.StringVar(&socks5yaml, "socks5conf", "socks5.yaml", "socks5代理配置")
}

var (
	socks5yaml  string
	config      Config
	upstreamMap map[string]*Upstream = make(map[string]*Upstream, 16)
)

type RouterRule struct {
	RuleType     string   `yaml:"rule-type"`
	Values       []string `yaml:"value,omitempty"`
	UpstreamName string   `yaml:"upstream-Name"`
}

func (receiver *RouterRule) determine(domain string, ip net.IP) bool {
	if receiver.RuleType == IP_CIDR && ip != nil {
		for _, value := range receiver.Values {
			_, ipNet, err := net.ParseCIDR(value)
			if err == nil && ipNet.Contains(ip) {
				return true
			}
		}
	} else if receiver.RuleType == DOMAIN_SUFFIX && domain != "" {
		for _, value := range receiver.Values {
			if strings.HasSuffix(domain, value) {
				return true
			}
		}
	} else if receiver.RuleType == MATCH {
		return true
	}
	return false
}

type Config struct {
	FinalUpstreamName string       `yaml:"final-upstream-name,omitempty"`
	Upstreams         []Upstream   `yaml:"upstreams"`
	GfwText           string       `yaml:"gfw-text,omitempty"`
	GfwUpstreamName   string       `yaml:"gfw-upstream-name,omitempty"`
	LocalAddr         string       `yaml:"local-addr"`
	Rules             []RouterRule `yaml:"rules"`
}

type Upstream struct {
	Name      string `yaml:"name"`
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	BasicAuth string `yaml:"basic-auth"`
}

func InfoUpstream(upstream *Upstream) string {
	if upstream == nil {
		return ""
	} else {
		return upstream.Host + ":" + strconv.Itoa(upstream.Port)
	}
}

// 修改无匹配的代理规则
func ModifyFinalUpstream(writer http.ResponseWriter, request *http.Request) {
	policy := request.URL.Query().Get("policy")
	log.Println("set final upstream as", policy)
	config.FinalUpstreamName = policy
	marshal, err := yaml.Marshal(config)
	if err != nil {
		writer.WriteHeader(500)
		writer.Header().Add("Content-Type", "text/text; charset=utf-8")
		writer.Write([]byte("error fetch config"))
	} else {
		writer.WriteHeader(200)
		writer.Header().Add("Content-Type", "text/text; charset=utf-8")
		writer.Write(marshal)
	}
}
