package list

import (
	"context"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/constants"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/core"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/gh"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/plgn"
	"github.com/urfave/cli/v3"
)

var github = &cli.Command{
	Name:  "github",
	Usage: "Lists all plugins that can be downloaded from Github",
	Action: func(ctx context.Context, c *cli.Command) error {
		release, err := gh.NewClient(core.Repo).GetReleaseByTag(ctx, constants.VersionWithPrefix())
		if err != nil {
			return core.ErrExit(err)
		}

		names := []string{}
		for _, asset := range release.Assets {
			if plgn.IsPluginReleaseAssetName(asset.Name) {
				names = append(names, asset.Name)
			}
		}

		if err := core.PrintResults(c, names); err != nil {
			return core.ErrExit(err)
		} else {
			return nil
		}
	},
}
