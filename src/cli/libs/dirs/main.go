package dirs

import (
	"os"
	"path/filepath"

	"github.com/chris-de-leon/chain-connectors/src/cli/libs/constants"
)

const APP_DIR = "chain-connectors-prototype"

func GetAppConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	} else {
		return filepath.Join(dir, APP_DIR, constants.VersionWithPrefix()), nil
	}
}

func GetAppCacheDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	} else {
		return filepath.Join(dir, APP_DIR, constants.VersionWithPrefix()), nil
	}
}

func GetAppGithubDir() (string, error) {
	dir, err := GetAppCacheDir()
	if err != nil {
		return "", err
	} else {
		return filepath.Join(dir, "github"), nil
	}
}

func GetAppPluginsDir() (string, error) {
	dir, err := GetAppConfigDir()
	if err != nil {
		return "", err
	} else {
		return filepath.Join(dir, "plugins"), nil
	}
}
