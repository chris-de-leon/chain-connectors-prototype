package install

import (
	"context"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/core"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/plgn"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"
)

var local = &cli.Command{
	Name:  "local",
	Usage: "Installs one or more plugins from a local file path",
	Flags: []cli.Flag{
		&cli.StringSliceFlag{Name: "plugin-path", Usage: "The path to the compiled plugin", Required: true},
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

		for _, pluginPath := range c.StringSlice("plugin-path") {
			eg.Go(func() error {
				pluginID := plgn.ID(pluginPath)

				if clean {
					if err := plgn.Store.Remove(pluginID); err != nil {
						return err
					}
				}

				return plgn.Store.Install([]string{pluginPath})
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
