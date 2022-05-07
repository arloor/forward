##  调试caddy-forwardproxy

## 在vps上部署forwardproxy

### 手动编译forwardproxy

> 需要使用go1.16,go 1.18不行

```shell
wget https://go.dev/dl/go1.16.15.linux-amd64.tar.gz -O go1.16.15.linux-amd64.tar.gz
rm -rf /usr/local/go16
mkdir /usr/local/go16
tar -zxvf go1.16.15.linux-amd64.tar.gz -C /usr/local/go16
ln -fs /usr/local/go16/go/bin/go /usr/local/bin/go16
go16 version
go16 install github.com/caddyserver/forwardproxy/cmd/caddy@8c6ef2bd4a8f40340b3ecd249f8eed058c567b76
```

### 下载二进制

```shell
mkdir /root/go
mkdir /root/go/bin
wget https://bwg.arloor.dev/caddy -O /root/go/bin/caddy
chmod +x /root/go/bin/caddy
```

### 设置服务

```shell
mkdir /root/go
mkdir /root/go/bin
wget https://bwg.arloor.dev/caddy -O /root/go/bin/caddy
chmod +x /root/go/bin/caddy

cat > /etc/caddyfile <<EOF
localhost:3128

forwardproxy {
    upstream         https://user:passwd@site:443
}
EOF

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

cat > /usr/local/bin/pass <<EOF
export http_proxy=http://localhost:3128
export https_proxy=http://localhost:3128
EOF

cat > /usr/local/bin/unpass <<EOF
export http_proxy=
export https_proxy=
EOF

. pass
curl https://google.com
```