package plgn

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/config"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/core"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/dirs"
)

type PluginStore struct {
	Dir string
}

var Store = PluginStore{dirs.PluginsConfig}

func (store *PluginStore) Install(filePaths []string) error {
	for _, filePath := range filePaths {
		dstDir := filepath.Join(store.Dir, ID(filePath))
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
	}

	return nil
}

func (store *PluginStore) IsInstalled(pluginID string) (bool, error) {
	_, err := store.GetPath(pluginID)
	if errors.Is(err, &PluginNotFoundError{}) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (store *PluginStore) GetPaths() ([]string, error) {
	entries, err := os.ReadDir(store.Dir)
	if os.IsNotExist(err) {
		return []string{}, nil
	}
	if err != nil {
		return []string{}, err
	}

	paths := make([]string, len(entries))
	for i, entry := range entries {
		paths[i] = filepath.Join(store.Dir, entry.Name())
	}
	return paths, nil
}

func (store *PluginStore) GetPath(pluginID string) (string, error) {
	pluginPaths, err := store.GetPaths()
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

	return "", &PluginNotFoundError{ids, pluginID}
}

func (store *PluginStore) Remove(pluginID string) error {
	pluginPath, err := store.GetPath(pluginID)
	if errors.Is(err, &PluginNotFoundError{}) {
		return nil
	}
	if err != nil {
		return err
	}
	return os.RemoveAll(filepath.Dir(pluginPath))
}

func (store *PluginStore) Exec(ctx context.Context, config *config.ChainConfig) error {
	pluginPath, err := store.GetPath(config.Plugin.ID)
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
