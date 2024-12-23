package common

import (
	"context"
	"fmt"

	"github.com/chris-de-leon/chain-connectors/src/cli/libs/constants"
	"github.com/urfave/cli/v3"
)

var version = &cli.Command{
	Name:  "version",
	Usage: "Returns the CLI version",
	Action: func(ctx context.Context, c *cli.Command) error {
		fmt.Fprintln(c.Root().Writer, constants.VersionWithPrefix())
		return nil
	},
}
