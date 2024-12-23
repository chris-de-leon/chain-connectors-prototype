package run

import (
	"context"

	"github.com/chris-de-leon/chain-connectors/src/cli/libs/config"
	"github.com/chris-de-leon/chain-connectors/src/cli/libs/core"
	"github.com/chris-de-leon/chain-connectors/src/cli/libs/plgn"
	"github.com/urfave/cli/v3"
)

var fromCLI = &cli.Command{
	Name:  "from-cli",
	Usage: "Run a plugin using the configurations passed in to this command",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "plugin-id", Usage: "The ID of the plugin to run", Sources: cli.EnvVars("PLUGIN_ID"), Required: true},
		&cli.StringFlag{Name: "server-host", Usage: "The server host", Sources: cli.EnvVars("SERVER_HOST"), Required: false, Value: "0.0.0.0"},
		&cli.IntFlag{Name: "server-port", Usage: "The server port", Sources: cli.EnvVars("SERVER_PORT"), Required: false, Value: 3000},
		&cli.StringFlag{Name: "chain-wss", Usage: "The chain WSS URL", Sources: cli.EnvVars("CHAIN_WSS_URL"), Required: false},
		&cli.StringFlag{Name: "chain-rpc", Usage: "The chain RPC URL", Sources: cli.EnvVars("CHAIN_RPC_URL"), Required: false},
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		if err := plgn.Exec(ctx,
			&config.ChainConfig{
				Plugin: &config.PluginConfig{
					ID: c.String("plugin-id"),
				},
				Server: &config.ServerConfig{
					Host: c.String("server-host"),
					Port: c.Int("server-port"),
				},
				Conn: &config.ConnectionConfg{
					Wss: c.String("chain-wss"),
					Rpc: c.String("chain-rpc"),
				},
			},
		); err != nil {
			return core.ErrExit(err)
		} else {
			return nil
		}
	},
}