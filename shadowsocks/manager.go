package shadowsocks

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"vpn-agent/utils"
)

type Config struct {
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
	LocalPort  int    `json:"local_port"`
	Password   string `json:"password"`
	Timeout    int    `json:"timeout"`
	FastOpen   bool   `json:"fast_open"`
	ReusePort  bool   `json:"reuse_port"`
	NoDelay    bool   `json:"no_delay"`
	Method     string `json:"method"`
	Plugin     string `json:"plugin"`
	PluginOpts string `json:"plugin_opts"`
}

type Manager struct {
}

func (m Manager) All() (interface{}, error) {
	return nil, nil
}

func (m Manager) Create() (interface{}, error) {
	port, err := utils.GetFreePort()
	if err != nil {
		return nil, err
	}

	serverHost, serverHostExists := os.LookupEnv("SERVER_HOST")
	if !serverHostExists {
		return nil, errors.New("SERVER_HOST ENV not found")
	}

	serverConfigHosts, serverConfigHostsExist := os.LookupEnv("SERVER_CONFIG_HOSTS")
	if !serverConfigHostsExist {
		return nil, errors.New("SERVER_CONFIG_HOSTS ENV not found")
	}

	clientPlugins, clientPluginsExist := os.LookupEnv("CLIENT_PLUGINS")
	if !clientPluginsExist {
		return nil, errors.New("CLIENT_PLUGINS ENV not found")
	}

	serverPlugins, serverPluginsExist := os.LookupEnv("SERVER_PLUGINS")
	if !serverPluginsExist {
		return nil, errors.New("SERVER_PLUGINS ENV not found")
	}

	clientPluginsOpts, clientPluginsOptsExist := os.LookupEnv("CLIENT_PLUGINS_OPTS")
	if !clientPluginsOptsExist {
		return nil, errors.New("CLIENT_PLUGINS ENV not found")
	}

	serverPluginsOpts, serverPluginsOptsExist := os.LookupEnv("SERVER_PLUGINS_OPTS")
	if !serverPluginsOptsExist {
		return nil, errors.New("SERVER_PLUGINS ENV not found")
	}

	password := utils.GeneratePassword(8)

	serverConfig := Config{
		Server:     strings.Trim(serverConfigHosts, "\\\""), //trim \"
		ServerPort: port,
		LocalPort:  1080,
		Password:   password,
		Timeout:    60,
		FastOpen:   false,
		ReusePort:  true,
		NoDelay:    true,
		Method:     "chacha20-ietf-poly1305",
		Plugin:     serverPlugins,
		PluginOpts: serverPluginsOpts,
	}

	clientConfig := serverConfig
	clientConfig.Server = serverHost
	clientConfig.Plugin = clientPlugins
	clientConfig.PluginOpts = clientPluginsOpts

	// open output file
	fo, err := os.Create("/etc/shadowsocks-libev/config_" + password + ".json")
	if err != nil {
		return nil, err
	}

	defer func() error {
		if err := fo.Close(); err != nil {
			return err
		}
		return nil
	}()

	data, err := json.Marshal(serverConfig)
	if err != nil {
		return nil, err
	}

	if _, err := fo.Write(data); err != nil {
		return nil, err
	}

	if err := m.enable(password); err != nil {
		return nil, fmt.Errorf("cannot enable service with password %s: %s", password, err.Error())
	}

	if err := m.start(password); err != nil {
		return nil, fmt.Errorf("cannot start service with password %s: %s", password, err.Error())
	}

	if err := m.status(password); err != nil {
		return nil, fmt.Errorf("status of service with password %s is not OK: %s", password, err.Error())
	}

	return clientConfig, nil
}

func (m Manager) Delete(password string) (interface{}, error) {
	if err := m.stop(password); err != nil {
		return nil, fmt.Errorf("cannot stop service with password %s: %s", password, err.Error())
	}

	if err := m.disable(password); err != nil {
		return nil, fmt.Errorf("cannot disable service with password %s: %s", password, err.Error())
	}

	// open output file
	err := os.Remove("/etc/shadowsocks-libev/config_" + password + ".json")
	if err != nil {
		return nil, err
	}

	return struct {
		Message string `json:"message"`
	}{
		Message: "Successfully deleted",
	}, nil
}

func (m Manager) enable(password string) error {
	return utils.ExecBash("systemctl enable --now shadowsocks-libev-server@" + password + ".service")
}

func (m Manager) disable(password string) error {
	return utils.ExecBash("systemctl disable --now shadowsocks-libev-server@" + password + ".service")
}

func (m Manager) start(password string) error {
	return utils.ExecBash("systemctl start shadowsocks-libev-server@" + password + ".service")
}

func (m Manager) stop(password string) error {
	return utils.ExecBash("systemctl stop shadowsocks-libev-server@" + password + ".service")
}

func (m Manager) status(password string) error {
	return utils.ExecBash("systemctl status shadowsocks-libev-server@" + password + ".service")
}
