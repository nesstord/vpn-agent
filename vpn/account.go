package vpn

import (
	"fmt"
	"os"
)

type Manager interface {
	All() (interface{}, error)
	Create() (interface{}, error)
	Delete(password string) (interface{}, error)
}

type Account struct {
	manager Manager
}

func (a Account) NewAccount() (Account, error) {
	protocol, exists := os.LookupEnv("VPN_PROTOCOL")
	if !exists {
		return Account{}, fmt.Errorf("env VPN_PROTOCOL does not exist")
	}

	m := AvailableAccountManagers[protocol]
	if m == nil {
		return Account{}, fmt.Errorf("manager for '%s' protocol not found", protocol)
	}

	return Account{
		manager: m,
	}, nil
}

func (a Account) All() (interface{}, error) {
	return a.manager.All()
}

func (a Account) Create() (interface{}, error) {
	return a.manager.Create()
}

func (a Account) Delete(password string) (interface{}, error) {
	return a.manager.Delete(password)
}
