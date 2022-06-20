**声明：本项目仅以学习为目的，请在当地法律允许的范围内使用本程序。任何因错误用途导致的法律责任，与本项目无关！**

## [HttpProxy](https://github.com/arloor/HttpProxy)的客户端

在本地启动http和socks5代理，作为[HttpProxy](https://github.com/arloor/HttpProxy)的客户端。socks5代理由自己开发，http代理则借鉴了caddy的forwardproxy插件。

![](/forward部署图.svg)

### 从Release页面下载二进制文件

- 提供了Windows64位和Linux64位可执行文件

### 配置文件

```shell
# http代理
cat > /etc/caddyfile <<EOF
localhost:3128

bind 127.0.0.1

forwardproxy {
    hide_ip
    hide_via
    upstream         socks5://localhost:1080 # 转发给本地的socks5代理
}
EOF

# socks5代理
cat > /etc/socks5.yaml <<EOF
upstream-alias:
  default: proxyA # 用于gfwlist
  final: direct   # 其余网站直连
upstreams:
  - name: proxyA
    host: xx.xx.xx.xx
    port: 443
    basic-auth: cHVxxxxxx3ZA== # user:passwd base64后的结果
gfw-text: E:\GoLandProjects\go-forward\gfwlist.txt
local-addr: localhost:1080
EOF
```

> socks5支持根据目标地址选择具体上游，具体配置例子见[socks5.yaml.example](/socks5.yaml.example)

### 运行指南

```shell
/root/go/bin/forward -socks5conf /etc/socks5.yaml -log /var/log/forward.log
```

### 测试

```shell
cat > /usr/local/bin/pass <<EOF
export http_proxy=http://localhost:3128
export https_proxy=http://localhost:3128
EOF

cat > /usr/local/bin/unpass <<EOF
export http_proxy=
export https_proxy=
EOF

. pass
curl https://www.google.com
```

### 其他feature

- 网速监控: [http://localhost:9999/](http://localhost:9999/)
- prometheus exporter: [http://localhost:9999/metrics](http://localhost:9999/metrics)
- 修改upstreamAlias：[http://localhost:9999/final?final=direct](http://localhost:9999/final?final=direct)