package config

import (
	"encoding/json"
	"os"

	"code.cloudfoundry.org/bbs/encryption"
	"code.cloudfoundry.org/lager/lagerflags"
)

type LocketConfig struct {
	DatabaseConnectionString string `json:"database_connection_string"`
	DatabaseDriver           string `json:"database_driver,omitempty"`
	ListenAddress            string `json:"listen_address"`
	encryption.EncryptionConfig
	lagerflags.LagerConfig
}

func DefaultLocketConfig() LocketConfig {
	return LocketConfig{
		LagerConfig:    lagerflags.DefaultLagerConfig(),
		DatabaseDriver: "mysql",
	}
}

func NewLocketConfig(configPath string) (LocketConfig, error) {
	locketConfig := DefaultLocketConfig()
	configFile, err := os.Open(configPath)
	if err != nil {
		return LocketConfig{}, err
	}
	decoder := json.NewDecoder(configFile)

	err = decoder.Decode(&locketConfig)
	if err != nil {
		return LocketConfig{}, err
	}

	return locketConfig, nil
}
