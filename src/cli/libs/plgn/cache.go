package plgn

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/core"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/dirs"
)

type PluginCache struct {
	Dir string
}

var Cache = PluginCache{dirs.PluginsCache}

func (cache *PluginCache) Download(ctx context.Context, pluginID string) ([]string, error) {
	cacheDir := filepath.Join(cache.Dir, pluginID)

	entries, err := os.ReadDir(cacheDir)
	if os.IsNotExist(err) {
		res, err := core.GithubClient.DownloadReleaseAsset(ctx, core.VersionWithPrefix(), MakePluginReleaseAssetName(pluginID))
		if err != nil {
			return []string{}, err
		} else {
			defer res.Body.Close()
		}
		return cache.unpack(res.Body, cacheDir)
	}
	if err != nil {
		return []string{}, err
	}

	pluginPaths := []string{}
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			pluginPaths = append(pluginPaths, filepath.Join(cacheDir, entry.Name()))
		}
	}

	return pluginPaths, nil
}

func (cache *PluginCache) unpack(src io.ReadCloser, dst string) ([]string, error) {
	pluginPaths := []string{}

	gzipReader, err := gzip.NewReader(src)
	if err != nil {
		return []string{}, err
	} else {
		defer gzipReader.Close()
	}

	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
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

		pth := filepath.Join(dst, header.Name)
		if header.Typeflag != tar.TypeReg {
			return []string{}, fmt.Errorf("failed to extract release asset(s) - not a regular file: %s", header.Name)
		}

		out, err := os.Create(pth)
		if err != nil {
			return []string{}, err
		} else {
			defer out.Close()
		}

		if _, err := io.Copy(out, tarReader); err != nil {
			return []string{}, err
		} else {
			pluginPaths = append(pluginPaths, pth)
		}
	}

	return pluginPaths, nil
}
