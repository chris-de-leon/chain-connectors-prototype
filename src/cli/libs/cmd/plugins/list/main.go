package list

import "github.com/urfave/cli/v3"

var Commands = &cli.Command{
	Name:  "list",
	Usage: "Commands for inspecting plugins",
	Commands: []*cli.Command{
		github,
		local,
	},
}
