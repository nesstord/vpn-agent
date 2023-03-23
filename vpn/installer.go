package vpn

type Installer interface {
	Commands() []string
}
