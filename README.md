## [HttpProxy](https://github.com/arloor/HttpProxy)的客户端

在本地启动http和socks5代理，作为[HttpProxy](https://github.com/arloor/HttpProxy)的客户端。socks5代理由自己开发，http代理使用caddy的forwardproxy插件。

![](/forward部署图.svg)

### 从Release页面下载二进制文件

- 目前最新版本为V4.0
- 提供了Windows64位和Linux64位可执行文件

### 从源码编译

> 需要使用go1.16，go 1.18不行

```shell
wget https://go.dev/dl/go1.16.15.linux-amd64.tar.gz -O go1.16.15.linux-amd64.tar.gz
rm -rf /usr/local/go16
mkdir /usr/local/go16
tar -zxvf go1.16.15.linux-amd64.tar.gz -C /usr/local/go16
ln -fs /usr/local/go16/go/bin/go /usr/local/bin/go16
go16 version
rm -rf forward
git clone https://github.com/arloor/forward
cd forward
go16 install forward/cmd/forward
```

### 配置文件

```shell
# http代理
cat > /etc/caddyfile <<EOF
localhost:3128

bind 127.0.0.1

forwardproxy {
    hide_ip
    hide_via
    # 上游地址，请修改为自己的
    upstream         https://user:passwd@proxy.site:443
}
EOF

# socks5代理
cat > /etc/socks5.yaml <<EOF
local-addr: localhost:1080 # 监听地址
upstreams:
  - name: default          # 上游名称
    host: proxy.site       # 上游地址
    port: 443              # 上游端口
    basic-auth: YXJsb2xxxxxxxxxxxxxvb3I= # "user:passwd" base64编码后的结果
rules:
  - rule-type: MATCH
    upstream-Name: default
EOF
```

> socks5支持根据目标地址选择具体上游，具体配置例子见[socks5.yaml.example](/socks5.yaml.example)

### 运行指南

```shell
/root/go/bin/forward -conf /etc/caddyfile -log /var/log/forward.log -socks5conf /etc/socks5.yaml
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
