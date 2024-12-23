package install

import (
	"github.com/urfave/cli/v3"
)

var Commands = &cli.Command{
	Name:  "install",
	Usage: "Commands for installing plugins",
	Commands: []*cli.Command{
		github,
		local,
	},
}
