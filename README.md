## [HttpProxy](https://github.com/arloor/HttpProxy)的客户端

在本地启动http和socks5代理，作为[HttpProxy](https://github.com/arloor/HttpProxy)的客户端。socks5代理由自己开发，http代理使用caddy的forwardproxy插件。

![](/forward部署图.svg)

### 从[releases](https://github.com/arloor/forward/releases/tag/V2.0)下载可执行文件

```shell
mkdir /root/go
mkdir /root/go/bin
wget https://github.com/arloor/forward/releases/download/V2.0/forward_linux_x64 -O /root/go/bin/forward
chmod +x /root/go/bin/forward
```

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
```

### linux设置开机启动

```shell
cat > /lib/systemd/system/con.service <<EOF
[Unit]
Description=本地http和socks5代理
After=network-online.target
Wants=network-online.target

[Service]
WorkingDirectory=/tmp
ExecStart=/root/go/bin/forward -conf /etc/caddyfile -log /var/log/forward.log -proxyHost proxy.site -proxyUser user -proxyPasswd passwd
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
