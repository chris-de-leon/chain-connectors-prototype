package dirs

import (
	"os"
	"path/filepath"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/core"
)

const (
	APPLICATION_DIR = "chain-connectors-prototype"
	PLUGINS_DIR     = "plugins"
)

var (
	PluginsConfig string
	PluginsCache  string
	Config        string
	Cache         string
)

func init() {
	var dir string
	var err error

	if dir, err = getConfigDir(); err != nil {
		panic(err)
	} else {
		Config = dir
	}

	if dir, err = getCacheDir(); err != nil {
		panic(err)
	} else {
		Cache = dir
	}

	if dir, err = getPluginsConfigDir(); err != nil {
		panic(err)
	} else {
		PluginsConfig = dir
	}

	if dir, err = getPluginsCacheDir(); err != nil {
		panic(err)
	} else {
		PluginsCache = dir
	}
}

func getConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	} else {
		return filepath.Join(dir, APPLICATION_DIR, core.VersionWithPrefix()), nil
	}
}

func getCacheDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	} else {
		return filepath.Join(dir, APPLICATION_DIR, core.VersionWithPrefix()), nil
	}
}

func getPluginsConfigDir() (string, error) {
	dir, err := getConfigDir()
	if err != nil {
		return "", err
	} else {
		return filepath.Join(dir, PLUGINS_DIR), nil
	}
}

func getPluginsCacheDir() (string, error) {
	dir, err := getCacheDir()
	if err != nil {
		return "", err
	} else {
		return filepath.Join(dir, PLUGINS_DIR), nil
	}
}
