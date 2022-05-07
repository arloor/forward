## [HttpProxy](https://github.com/arloor/HttpProxy)的客户端

直接使用caddy的forwardproxy插件作为我的https代理的客户端软件，利用其upstream能力将客户端代理转发到我的HttpProxy上

### 获取二进制可执行文件

可以从源码编译，也可以直接下载可执行文件

**1. 从源码编译**

> 需要使用go1.16，go1.18不行

```shell
wget https://go.dev/dl/go1.16.15.linux-amd64.tar.gz -O go1.16.15.linux-amd64.tar.gz
rm -rf /usr/local/go16
mkdir /usr/local/go16
tar -zxvf go1.16.15.linux-amd64.tar.gz -C /usr/local/go16
ln -fs /usr/local/go16/go/bin/go /usr/local/bin/go16
go16 version
go16 install github.com/caddyserver/forwardproxy/cmd/caddy@8c6ef2bd4a8f40340b3ecd249f8eed058c567b76
```

**2. 下载可执行文件**

```shell
mkdir /root/go
mkdir /root/go/bin
wget https://github.com/arloor/forward/releases/download/v1.0/caddy-linux-amd64 -O /root/go/bin/caddy
chmod +x /root/go/bin/caddy
```
### 配置文件

```shell
cat > /etc/caddyfile <<EOF
localhost:3128

bind 127.0.0.1

forwardproxy {
    hide_ip
    hide_via
    # 上游地址
    upstream         https://user:passwd@site:443
}
EOF
```

### linux设置开机启动

```shell
cat > /lib/systemd/system/con.service <<EOF
[Unit]
Description=forwardproxy-Http代理
After=network-online.target
Wants=network-online.target

[Service]
WorkingDirectory=/tmp
ExecStart=/root/go/bin/caddy -conf /etc/caddyfile
LimitNOFILE=100000
Restart=always
RestartSec=30

[Install]
WantedBy=multi-user.target
EOF
service con stop
systemctl daemon-reload
service con start
service con status
systemctl enable  con
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
