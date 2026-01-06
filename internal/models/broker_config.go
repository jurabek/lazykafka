package models

type AuthType int

const (
	AuthNone AuthType = iota
	AuthSASL
)

func (a AuthType) String() string {
	if a == AuthSASL {
		return "SASL"
	}
	return "None"
}

type SASLMechanism int

const (
	SASLPlain SASLMechanism = iota
	SASLSCRAMSHA256
	SASLSCRAMSHA512
	SASLOAuthBearer
)

func (s SASLMechanism) String() string {
	switch s {
	case SASLSCRAMSHA256:
		return "SCRAM-SHA-256"
	case SASLSCRAMSHA512:
		return "SCRAM-SHA-512"
	case SASLOAuthBearer:
		return "OAUTHBEARER"
	default:
		return "PLAIN"
	}
}

type BrokerConfig struct {
	Name             string        `json:"name"`
	BootstrapServers string        `json:"bootstrap_servers"`
	AuthType         AuthType      `json:"auth_type"`
	SASLMechanism    SASLMechanism `json:"sasl_mechanism,omitempty"`
	Username         string        `json:"username,omitempty"`
	Password         string        `json:"password,omitempty"`
}
