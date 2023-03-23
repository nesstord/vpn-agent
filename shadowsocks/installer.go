package shadowsocks

type Installer struct {
}

func (i Installer) Commands() []string {
	return []string{
		`apt -y update
		apt install -y --allow-remove-essential kmod wget cron iproute2 shadowsocks-libev fail2ban

		SERVER_HOST=$(ip addr | grep 'inet' | grep -v inet6 | grep -vE '127\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | grep -oE '[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' | head -1)
		SERVER_CONFIG_HOSTS="\"${SERVER_HOST}\"" # ipv4

		echo '[INCLUDES]' > /etc/fail2ban/filter.d/shadowsocks.conf
		echo 'before = common.conf' >> /etc/fail2ban/filter.d/shadowsocks.conf
		echo '[Definition]' >> /etc/fail2ban/filter.d/shadowsocks.conf
		echo '_daemon = ss-server' >> /etc/fail2ban/filter.d/shadowsocks.conf
		echo 'failregex = ^%(__prefix_line)s.*ERROR: failed to handshake with <HOST>: (:?authentication error$|malicious fragmentation$)' >> /etc/fail2ban/filter.d/shadowsocks.conf
		echo 'ignoreregex =' >> /etc/fail2ban/filter.d/shadowsocks.conf

		echo '[shadowsocks]' > /etc/fail2ban/jail.local
		echo 'enabled = true' >> /etc/fail2ban/jail.local
		echo 'port    = 8388' >> /etc/fail2ban/jail.local
		echo 'logpath  = /var/log/syslog' >> /etc/fail2ban/jail.local
		echo 'maxretry = 3' >> /etc/fail2ban/jail.local
		echo 'bantime = -1' >> /etc/fail2ban/jail.local
		echo 'findtime = 5' >> /etc/fail2ban/jail.local

		PLUGIN_V2RAY_VERSION="1.3.1"
		PLUGIN_V2RAY_RELEASE="https://github.com/shadowsocks/v2ray-plugin/releases/download/v${PLUGIN_V2RAY_VERSION}/v2ray-plugin-linux-amd64-v${PLUGIN_V2RAY_VERSION}.tar.gz"

		echo "Setup V2RAY Plugin"
		wget -qO- "${PLUGIN_V2RAY_RELEASE}" | tar xvz -C /etc/shadowsocks-libev/ || exit 1
		SERVER_PLUGINS="/etc/shadowsocks-libev/v2ray-plugin_linux_amd64"
		SERVER_PLUGINS_OPTS="server"
		CLIENT_PLUGINS="/etc/shadowsocks-libev/v2ray-plugin_linux_amd64"
		CLIENT_PLUGINS_OPTS="host=google.com"

		echo "apt-get -y update
		apt-get -y --only-upgrade install shadowsocks-libev
		/etc/init.d/shadowsocks-libev restart" > /shadowsocks-update.sh
		chmod +x /shadowsocks-update.sh
		crontab -l | { cat; echo "0 0 * * 0 bash /bin/bash -c \"/shadowsocks-update.sh\""; } | crontab -
		service cron restart

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

		AUTH_TOKEN=$(echo $RANDOM | md5sum | head -c 24; echo)

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
		Environment=AUTH_TOKEN=${AUTH_TOKEN}
		StandardOutput=syslog
		StandardError=syslog
		SyslogIdentifier=vpn-agent
		[Install]
		WantedBy=multi-user.target
		EOF

		echo "VPN-Agent service successfully installed"

		systemctl daemon-reload
		systemctl enable vpn-agent
		systemctl start vpn-agent
		sleep 5
		systemctl status vpn-agent

		echo "Use this token $AUTH_TOKEN to authenticate"`,
	}
}
