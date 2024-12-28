package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func ParseCliConfig(filePath string) (*CliConfig, error) {
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

func ParseChainConfig(filePath string, chainName string) (*ChainConfig, error) {
	cliConfig, err := ParseCliConfig(filePath)
	if err != nil {
		return nil, err
	}

	chainConfig, exists := cliConfig.Chains[chainName]
	if !exists {
		return nil, fmt.Errorf(
			"chain with name '%s' does not exist in config - must be one of: [ %s ]",
			chainName,
			strings.Join(cliConfig.ChainNames(), ", "),
		)
	}

	return &chainConfig, nil
}
