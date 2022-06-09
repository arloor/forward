module forward

require (
	github.com/caddyserver/caddy v1.0.5
	github.com/caddyserver/forwardproxy v0.0.0-20211013034647-8c6ef2bd4a8f
	//github.com/prometheus/client_golang v1.12.2-0.20220603061532-0dd939295e75
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/caddyserver/forwardproxy v0.0.0-20211013034647-8c6ef2bd4a8f => github.com/arloor/forwardproxy v0.0.0-20220528070509-f3c1b36e0a33

go 1.16
