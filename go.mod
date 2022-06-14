module forward

require (
	github.com/BurntSushi/toml v1.1.0 // indirect
	github.com/caddyserver/forwardproxy v0.0.0-20211013034647-8c6ef2bd4a8f
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/caddyserver/forwardproxy v0.0.0-20211013034647-8c6ef2bd4a8f => github.com/arloor/forwardproxy v0.0.0-20220614075435-ccf74f60544c

go 1.16
