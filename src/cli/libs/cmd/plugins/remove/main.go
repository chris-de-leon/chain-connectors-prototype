package remove

import (
	"context"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/core"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/plgn"
	"github.com/urfave/cli/v3"
)

var Commands = &cli.Command{
	Name:  "remove",
	Usage: "Removes a locally installed plugin",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{Name: "plugin-id", Usage: "The ID of the plugin to remove", Required: false},
		&cli.BoolFlag{Name: "all", Usage: "Remove all plugins", Aliases: []string{"a"}, Required: false},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		pluginIDs := c.StringSlice("plugin-id")

		if c.Bool("all") {
			if ids, err := plgn.IDs(); err != nil {
				return core.ErrExit(err)
			} else {
				pluginIDs = ids
			}
		}

		for _, pluginID := range pluginIDs {
			if err := plgn.Store.Remove(pluginID); err != nil {
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
