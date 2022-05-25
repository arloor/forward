module forward

require (
	github.com/caddyserver/caddy v1.0.5
	github.com/caddyserver/forwardproxy v0.0.0-20211013034647-8c6ef2bd4a8f
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/caddyserver/forwardproxy v0.0.0-20211013034647-8c6ef2bd4a8f => github.com/arloor/forwardproxy v0.0.0-20220525105858-6abe7701be35

go 1.16
