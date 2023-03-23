package vpn

import (
	"vpn-agent/shadowsocks"
)

var AvailableInstallers = map[string]Installer{
	"shadowsocks": shadowsocks.Installer{},
}

var AvailableAccountManagers = map[string]Manager{
	"shadowsocks": shadowsocks.Manager{},
}
