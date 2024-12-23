package plugins

import (
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/cmd/plugins/install"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/cmd/plugins/list"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/cmd/plugins/remove"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/cmd/plugins/run"
	"github.com/urfave/cli/v3"
)

var Commands = &cli.Command{
	Name:  "plugins",
	Usage: "Commands for managing plugins",
	Commands: []*cli.Command{
		install.Commands,
		remove.Commands,
		list.Commands,
		run.Commands,
	},
}
