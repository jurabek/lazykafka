package secrets

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
)

const serviceName = "lazykafka"

type SecretStore interface {
	SaveCredentials(brokerName, username, password string) error
	GetCredentials(brokerName string) (username, password string, err error)
	DeleteCredentials(brokerName string) error
}

type KeyringStore struct{}

func NewKeyringStore() *KeyringStore {
	return &KeyringStore{}
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (k *KeyringStore) SaveCredentials(brokerName, username, password string) error {
	creds := credentials{
		Username: username,
		Password: password,
	}
	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("marshal credentials: %w", err)
	}
	return keyring.Set(serviceName, brokerName, string(data))
}

func (k *KeyringStore) GetCredentials(brokerName string) (string, string, error) {
	data, err := keyring.Get(serviceName, brokerName)
	if err != nil {
		return "", "", err
	}
	var creds credentials
	if err := json.Unmarshal([]byte(data), &creds); err != nil {
		return "", "", fmt.Errorf("unmarshal credentials: %w", err)
	}
	return creds.Username, creds.Password, nil
}

func (k *KeyringStore) DeleteCredentials(brokerName string) error {
	return keyring.Delete(serviceName, brokerName)
}
