package models

type AuthType int

const (
	AuthNone AuthType = iota
	AuthSASL
	AuthSSL
)

func (a AuthType) String() string {
	switch a {
	case AuthSASL:
		return "SASL"
	case AuthSSL:
		return "SSL"
	default:
		return "None"
	}
}

type BrokerConfig struct {
	Name             string   `json:"name"`
	BootstrapServers string   `json:"bootstrap_servers"`
	AuthType         AuthType `json:"auth_type"`
	Username         string   `json:"username,omitempty"`
	Password         string   `json:"password,omitempty"`
}
