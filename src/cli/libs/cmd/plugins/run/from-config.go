package run

import (
	"context"
	"fmt"
	"strings"

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

		cliConfig, err := config.Parse(configPath)
		if err != nil {
			return core.ErrExit(err)
		}

		chainConfig, exists := cliConfig.Chains[chainName]
		if !exists {
			return fmt.Errorf(
				"chain with name '%s' does not exist in CLI config - must be one of: [ %s ]",
				chainName,
				strings.Join(cliConfig.ChainNames(), ", "),
			)
		}

		if err := plgn.Exec(ctx, &chainConfig); err != nil {
			return core.ErrExit(err)
		} else {
			return nil
		}
	},
}
