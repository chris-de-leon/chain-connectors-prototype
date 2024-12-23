package common

import (
	"context"
	"path/filepath"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/core"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/dirs"
	"github.com/urfave/cli/v3"
)

var clean = &cli.Command{
	Name:  "clean",
	Usage: "Cleans up the CLI config directory and the CLI cache directory",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "config", Usage: "If specified, remove all data from the CLI config directory", Required: false},
		&cli.BoolFlag{Name: "cache", Usage: "If specified, remove all data from the CLI cache directory", Required: false},
		&cli.BoolFlag{Name: "force", Usage: "If specified, skip all prompts", Aliases: []string{"f"}, Required: false, Value: false},
		&cli.BoolFlag{Name: "all", Usage: "If specified, remove all data from prior CLI versions", Aliases: []string{"a"}, Required: false, Value: false},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		config := c.Bool("config")
		cache := c.Bool("cache")
		force := c.Bool("force")
		all := c.Bool("all")

		if !config && !cache {
			config = true
			cache = true
		}

		if config {
			configDir, err := dirs.GetAppConfigDir()
			if err != nil {
				return core.ErrExit(err)
			}

			if all {
				configDir = filepath.Dir(configDir)
			}

			if err := core.CleanDir(c, configDir, force); err != nil {
				return core.ErrExit(err)
			}
		}

		if cache {
			cacheDir, err := dirs.GetAppCacheDir()
			if err != nil {
				return core.ErrExit(err)
			}

			if all {
				cacheDir = filepath.Dir(cacheDir)
			}

			if err := core.CleanDir(c, cacheDir, force); err != nil {
				return core.ErrExit(err)
			}
		}

		return nil
	},
}
