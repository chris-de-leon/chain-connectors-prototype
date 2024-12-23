package list

import (
	"context"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/core"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/plgn"
	"github.com/urfave/cli/v3"
)

var local = &cli.Command{
	Name:  "local",
	Usage: "Lists the IDs of all locally installed plugins",
	Action: func(ctx context.Context, c *cli.Command) error {
		ids, err := plgn.IDs()
		if err != nil {
			return core.ErrExit(err)
		}

		if err := core.PrintResults(c, ids); err != nil {
			return core.ErrExit(err)
		} else {
			return nil
		}
	},
}
