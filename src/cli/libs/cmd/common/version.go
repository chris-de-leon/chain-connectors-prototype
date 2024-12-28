package common

import (
	"context"
	"fmt"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/core"
	"github.com/urfave/cli/v3"
)

var version = &cli.Command{
	Name:  "version",
	Usage: "Returns the CLI version",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "no-prefix", Usage: "If specified remove the leading 'v'", Required: false, Value: false},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		if c.Bool("no-prefix") {
			fmt.Fprintln(c.Root().Writer, core.VersionWithoutPrefix())
		} else {
			fmt.Fprintln(c.Root().Writer, core.VersionWithPrefix())
		}
		return nil
	},
}
