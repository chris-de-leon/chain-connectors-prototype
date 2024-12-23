package config

import (
	"encoding/json"
	"os"
)

func Parse(filePath string) (*CliConfig, error) {
	var conf CliConfig

	configFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	if err := json.NewDecoder(configFile).Decode(&conf); err != nil {
		return nil, err
	}

	return &conf, err
}
