package data

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/jurabek/lazykafka/internal/models"
)

type BrokerStorage interface {
	Load() ([]models.BrokerConfig, error)
	Save(configs []models.BrokerConfig) error
}

type FileBrokerStorage struct {
	filePath string
}

func NewFileBrokerStorage() (*FileBrokerStorage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".lazykafka")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	return &FileBrokerStorage{
		filePath: filepath.Join(configDir, "brokers.json"),
	}, nil
}

func (s *FileBrokerStorage) Load() ([]models.BrokerConfig, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.BrokerConfig{}, nil
		}
		return nil, err
	}

	var configs []models.BrokerConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, err
	}

	return configs, nil
}

func (s *FileBrokerStorage) Save(configs []models.BrokerConfig) error {
	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0600)
}
