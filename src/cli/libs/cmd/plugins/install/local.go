package install

import (
	"context"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/core"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/plgn"
	"github.com/urfave/cli/v3"
)

var local = &cli.Command{
	Name:  "local",
	Usage: "Installs one or more plugins from a local file path",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{Name: "plugin-path", Usage: "The path to the compiled plugin", Required: true},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		for _, pluginPath := range c.StringSlice("plugin-path") {
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
