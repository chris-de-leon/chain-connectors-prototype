package run

import "github.com/urfave/cli/v3"

var Commands = &cli.Command{
	Name:  "run",
	Usage: "Commands for running plugins",
	Commands: []*cli.Command{
		fromConfig,
		fromCLI,
	},
}
