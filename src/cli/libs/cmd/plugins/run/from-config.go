package run

import (
	"context"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/config"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/core"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/plgn"
	"github.com/urfave/cli/v3"
)

var fromConfig = &cli.Command{
	Name:  "from-config",
	Usage: "Run a plugin using the configurations defined in a config file",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "config", Usage: "The path to the CLI config file", Aliases: []string{"c"}, Sources: cli.EnvVars("CONFIG"), Required: true},
		&cli.StringFlag{Name: "name", Usage: "The name of the chain", Aliases: []string{"n"}, Sources: cli.EnvVars("CHAIN"), Required: true},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		configPath := c.String("config")
		chainName := c.String("name")

		chainConfig, err := config.ParseChainConfig(configPath, chainName)
		if err != nil {
			return core.ErrExit(err)
		}

		isInstalled, err := plgn.Store.IsInstalled(chainConfig.Plugin.ID)
		if err != nil {
			return core.ErrExit(err)
		}

		if !isInstalled {
			pluginPaths, err := plgn.Cache.Download(ctx, chainConfig.Plugin.ID)
			if err != nil {
				return core.ErrExit(err)
			}

			if err := plgn.Store.Install(pluginPaths); err != nil {
				return core.ErrExit(err)
			}
		}

		if err := plgn.Store.Exec(ctx, chainConfig); err != nil {
			return core.ErrExit(err)
		} else {
			return nil
		}
	},
}
