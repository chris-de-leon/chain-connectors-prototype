package install

import (
	"context"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/core"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/plgn"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

var github = &cli.Command{
	Name:  "github",
	Usage: "Installs a plugin from github",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{Name: "plugin-id", Usage: "The ID of the plugin to install", Required: true},
		&cli.BoolFlag{Name: "clean", Usage: "If the plugin already exists, then remove it and re-install it", Required: false, Value: false},
		&cli.IntFlag{Name: "concurrency", Usage: "The maximum number of concurrent requests", Required: false, Value: 0},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		concurrency := c.Int("concurrency")
		clean := c.Bool("clean")

		eg := new(errgroup.Group)
		if concurrency > 0 {
			eg.SetLimit(int(concurrency))
		}

		for _, pluginID := range c.StringSlice("plugin-id") {
			eg.Go(func() error {
				if clean {
					if err := plgn.Store.Remove(pluginID); err != nil {
						return err
					}
				}

				pluginPaths, err := plgn.Cache.Download(ctx, pluginID)
				if err != nil {
					return err
				} else {
					return plgn.Store.Install(pluginPaths)
				}
			})
		}

		if err := eg.Wait(); err != nil {
			return core.ErrExit(err)
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
