#!/bin/bash

#
apt update
apt install -y kmod wget cron systemctl iproute2

DNS_PRIMARY="${DNS_PRIMARY:-208.67.222.222}" # default: opendns
DNS_SECONDARY="${DNS_SECONDARY:-208.67.220.220}" # default: opendns
SERVER_HOST=$(ip addr | grep 'inet' | grep -v inet6 | grep -vE '127\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | grep -oE '[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | head -1)
SERVER_IPV6_HOST=$(ip -6 addr | grep inet6 | awk -F '[ \t]+|/' '{print $3}' | grep -v ^::1 | grep -v ^fe80)

PLUGIN_V2RAY_VERSION="1.3.1"
PLUGIN_V2RAY_RELEASE="https://github.com/shadowsocks/v2ray-plugin/releases/download/v${PLUGIN_V2RAY_VERSION}/v2ray-plugin-linux-amd64-v${PLUGIN_V2RAY_VERSION}.tar.gz"
SHADOWSOCKS_PLUGIN_V2RAY_ENABLE="1"
VPN_CLIENT_CONFIG_FILE="${VPN_CLIENT_CONFIG_FILE:-/tmp/shadowsocks-client-config}"


SERVER_CONFIG_HOSTS="\"${SERVER_HOST}\"" # ipv4
if [ "$SERVER_IPV6_HOST" ]
then
    SERVER_CONFIG_HOSTS="[\"${SERVER_IPV6_HOST}\",\"${SERVER_HOST}\"]" # ipv4 & ipv6
fi

echo "nameserver $DNS_PRIMARY
nameserver $DNS_SECONDARY" > /etc/resolv.conf

# They optimize the server networking protocols
echo '* soft nofile 51200' >> /etc/security/limits.conf
echo '* hard nofile 51200' >> /etc/security/limits.conf
ulimit -n 51200
echo 'fs.file-max = 51200
net.core.rmem_max = 67108864
net.core.wmem_max = 67108864
net.core.netdev_max_backlog = 250000
net.core.somaxconn = 4096
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 1200
net.ipv4.ip_local_port_range = 10000 65000
net.ipv4.tcp_max_syn_backlog = 8192
net.ipv4.tcp_max_tw_buckets = 5000
net.ipv4.tcp_fastopen = 3
net.ipv4.tcp_mem = 25600 51200 102400
net.ipv4.tcp_rmem = 4096 87380 67108864
net.ipv4.tcp_wmem = 4096 65536 67108864
net.ipv4.tcp_mtu_probing = 1
net.ipv4.tcp_congestion_control = hybla' > /etc/sysctl.conf
sysctl -p

modprobe tcp_bbr
sh -c 'echo "tcp_bbr" >> /etc/modules-load.d/modules.conf'
sh -c 'echo "net.core.default_qdisc=fq" >> /etc/sysctl.conf'
sh -c 'echo "net.ipv4.tcp_congestion_control=bbr" >> /etc/sysctl.conf'
lsmod | grep bbr

sysctl -p

if ! [ -x "$(command -v sudo)" ]; then
  apt update -y && apt install -y sudo
fi

apt -y update &&
apt install -y shadowsocks-libev fail2ban || exit 1

# Fail2ban
cat > /etc/fail2ban/filter.d/shadowsocks.conf<<-EOF
[INCLUDES]
before = common.conf
[Definition]
_daemon = ss-server
failregex = ^%(__prefix_line)s.*ERROR: failed to handshake with <HOST>: (:?authentication error$|malicious fragmentation$)
ignoreregex =
EOF

cat > /etc/fail2ban/jail.local<<-EOF
[shadowsocks]
enabled = true
port    = 8388
logpath  = /var/log/syslog
maxretry = 3
bantime = -1
findtime = 5
EOF
systemctl restart fail2ban

# V2RAY
SERVER_PLUGINS=""
SERVER_PLUGINS_OPTS=""
CLIENT_PLUGINS=""
CLIENT_PLUGINS_OPTS=""

if [ "$SHADOWSOCKS_PLUGIN_V2RAY_ENABLE" == "1" ]; then
    echo "Setup V2RAY Plugin"
    wget -qO- $PLUGIN_V2RAY_RELEASE | sudo tar xvz -C /etc/shadowsocks-libev/ || exit 1
    SERVER_PLUGINS="/etc/shadowsocks-libev/v2ray-plugin_linux_amd64"
    SERVER_PLUGINS_OPTS="server"
    CLIENT_PLUGINS="/etc/shadowsocks-libev/v2ray-plugin_linux_amd64"
    CLIENT_PLUGINS_OPTS="host=google.com"
fi

#sed -i '$ d' $VPN_CLIENT_CONFIG_FILE # remove last line (next account config)
touch $VPN_CLIENT_CONFIG_FILE

# shadowsocks autoupdate
echo "apt-get -y update
apt-get -y --only-upgrade install shadowsocks-libev
/etc/init.d/shadowsocks-libev restart" > /shadowsocks-update.sh
chmod +x /shadowsocks-update.sh
crontab -l | { cat; echo "0 0 * * 0 bash /bin/bash -c \"/shadowsocks-update.sh\""; } | crontab -
service cron restart

# create daemon
cat > "/lib/systemd/system/shadowsocks-libev-server@.service"<<-EOF
[Unit]
Description=Shadowsocks-Libev Custom Server Service for %I
Documentation=man:ss-server(1)
After=network-online.target

[Service]
Type=simple
CapabilityBoundingSet=CAP_NET_BIND_SERVICE
ExecStart=/usr/bin/ss-server -c /etc/shadowsocks-libev/config_%i.json
LimitNOFILE=32768

[Install]
WantedBy=multi-user.target
EOF

/etc/init.d/shadowsocks-libev stop

echo "Shadowsocks server successfully installed"

# create vpn-agent daemon
cat > "/lib/systemd/system/vpn-agent.service"<<-EOF
[Unit]
Description=VPN-Agent service

[Service]
Type=simple
User=root
WorkingDirectory=/root
ExecStart=/usr/local/bin/vpn-agent run
Restart=always
RestartSec=5s
Environment=VPN_PROTOCOL=shadowsocks
Environment=SERVER_CONFIG_HOSTS=${SERVER_CONFIG_HOSTS}
Environment=SERVER_HOST=${SERVER_HOST}
Environment=CLIENT_PLUGINS=${CLIENT_PLUGINS}
Environment=SERVER_PLUGINS=${SERVER_PLUGINS}
Environment=CLIENT_PLUGINS_OPTS=${CLIENT_PLUGINS_OPTS}
Environment=SERVER_PLUGINS_OPTS=${SERVER_PLUGINS_OPTS}
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=vpn-agent

[Install]
WantedBy=multi-user.target
EOF

echo "VPN-Agent service successfully installed"

#check daemons
systemctl daemon-reload
systemctl enable vpn-agent
systemctl start vpn-agent
sleep 5
systemctl status vpn-agent