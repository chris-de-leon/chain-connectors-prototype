package plgn

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/chris-de-leon/chain-connectors/src/cli/libs/config"
	"github.com/chris-de-leon/chain-connectors/src/cli/libs/constants"
	"github.com/chris-de-leon/chain-connectors/src/cli/libs/core"
	"github.com/chris-de-leon/chain-connectors/src/cli/libs/dirs"
)

func IsPluginReleaseAssetName(name string) bool {
	return strings.HasSuffix(
		name,
		fmt.Sprintf(
			"-plugin_%s_%s_%s.tar.gz",
			constants.VersionWithoutPrefix(),
			runtime.GOOS,
			runtime.GOARCH,
		),
	)
}

func MakePluginReleaseAssetName(pluginID string) string {
	return fmt.Sprintf(
		"%s-plugin_%s_%s_%s.tar.gz",
		pluginID,
		constants.VersionWithoutPrefix(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}

func Install(filePath string) error {
	plgDir, err := dirs.GetAppPluginsDir()
	if err != nil {
		return err
	}

	dstDir := filepath.Join(plgDir, ID(filePath))
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return err
	}

	dstPlg := filepath.Join(dstDir, "bin")
	if err := os.Link(filePath, dstPlg); err != nil {
		return err
	}

	if err := os.Chmod(dstPlg, core.FileModeExecutable); err != nil {
		return err
	}

	return nil
}

func Remove(pluginID string) error {
	pluginPath, err := GetPath(pluginID)
	if err != nil {
		return err
	} else {
		return os.RemoveAll(filepath.Dir(pluginPath))
	}
}

func GetPaths() ([]string, error) {
	plgDir, err := dirs.GetAppPluginsDir()
	if err != nil {
		return []string{}, err
	}

	entries, err := os.ReadDir(plgDir)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return []string{}, err
	}

	paths := make([]string, len(entries))
	for i, entry := range entries {
		paths[i] = filepath.Join(plgDir, entry.Name())
	}
	return paths, nil
}

func GetPath(pluginID string) (string, error) {
	pluginPaths, err := GetPaths()
	if err != nil {
		return "", err
	}

	ids := make([]string, len(pluginPaths))
	for i, pluginPath := range pluginPaths {
		id := ID(pluginPath)
		if id == pluginID {
			return filepath.Join(pluginPath, "bin"), nil
		} else {
			ids[i] = id
		}
	}

	return "", fmt.Errorf(
		"plugin with ID '%s' does not exist - must be one of: [ %s ]",
		pluginID,
		strings.Join(ids, ", "),
	)
}

func ID(filePath string) string {
	return filepath.Base(filePath)
}

func IDs() ([]string, error) {
	pluginPaths, err := GetPaths()
	if err != nil {
		return []string{}, err
	}

	pluginIDs := make([]string, len(pluginPaths))
	for i, p := range pluginPaths {
		pluginIDs[i] = ID(p)
	}

	return pluginIDs, nil
}

func Exec(ctx context.Context, config *config.ChainConfig) error {
	pluginPath, err := GetPath(config.Plugin.ID)
	if err != nil {
		return err
	}

	confBytes, err := json.Marshal(config)
	if err != nil {
		return err
	}

	cmd := exec.CommandContext(ctx, pluginPath, string(confBytes))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func Unpack(pluginID string, src io.ReadCloser) ([]string, error) {
	pluginPaths := []string{}

	gzipReader, err := gzip.NewReader(src)
	if err != nil {
		return []string{}, err
	} else {
		defer gzipReader.Close()
	}

	ghDir, err := dirs.GetAppGithubDir()
	if err != nil {
		return []string{}, err
	}

	dstDir := filepath.Join(ghDir, pluginID)
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return []string{}, err
	}

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return []string{}, err
		}

		dst := filepath.Join(dstDir, header.Name)
		if header.Typeflag != tar.TypeReg {
			return []string{}, fmt.Errorf("failed to extract release asset(s) - not a regular file: %s", header.Name)
		}

		out, err := os.Create(dst)
		if err != nil {
			return []string{}, err
		} else {
			defer out.Close()
		}

		if _, err := io.Copy(out, tarReader); err != nil {
			return []string{}, err
		} else {
			pluginPaths = append(pluginPaths, dst)
		}
	}

	return pluginPaths, nil
}
