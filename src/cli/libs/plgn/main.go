package plgn

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/core"
)

func IsPluginReleaseAssetName(name string) bool {
	return strings.HasSuffix(
		name,
		fmt.Sprintf(
			"-plugin_%s_%s_%s.tar.gz",
			core.VersionWithoutPrefix(),
			runtime.GOOS,
			runtime.GOARCH,
		),
	)
}

func MakePluginReleaseAssetName(pluginID string) string {
	return fmt.Sprintf(
		"%s-plugin_%s_%s_%s.tar.gz",
		pluginID,
		core.VersionWithoutPrefix(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}

func ID(filePath string) string {
	return filepath.Base(filePath)
}

func IDs() ([]string, error) {
	pluginPaths, err := Store.GetPaths()
	if err != nil {
		return []string{}, err
	}

	pluginIDs := make([]string, len(pluginPaths))
	for i, p := range pluginPaths {
		pluginIDs[i] = ID(p)
	}

	return pluginIDs, nil
}
