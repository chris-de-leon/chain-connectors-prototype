package install

import (
	"context"

	"github.com/chris-de-leon/chain-connectors/src/cli/libs/constants"
	"github.com/chris-de-leon/chain-connectors/src/cli/libs/core"
	"github.com/chris-de-leon/chain-connectors/src/cli/libs/gh"
	"github.com/chris-de-leon/chain-connectors/src/cli/libs/plgn"
	"github.com/urfave/cli/v3"
)

var github = &cli.Command{
	Name:  "github",
	Usage: "Installs a plugin from github",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "plugin-id", Usage: "The ID of the plugin to install", Required: true},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		pluginID := c.String("plugin-id")

		res, err := gh.NewClient(core.Repo).DownloadReleaseAsset(ctx, constants.VersionWithPrefix(), plgn.MakePluginReleaseAssetName(pluginID))
		if err != nil {
			return core.ErrExit(err)
		} else {
			defer res.Body.Close()
		}

		pluginPaths, err := plgn.Unpack(pluginID, res.Body)
		if err != nil {
			return core.ErrExit(err)
		}

		for _, pluginPath := range pluginPaths {
			if err := plgn.Install(pluginPath); err != nil {
				return core.ErrExit(err)
			}
		}

		ids, err := plgn.IDs()
		if err != nil {
			return core.ErrExit(err)
		}

		if err := core.PrintResults(c, ids); err != nil {
			return core.ErrExit(err)
		}

		return nil
	},
}
