package cmd

import (
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/cmd/common"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/cmd/plugins"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/constants"
	"github.com/urfave/cli/v3"
)

var Commands = &cli.Command{
	Name:    "Chain Connectors",
	Usage:   "CLI",
	Version: constants.VersionWithPrefix(),
	Commands: append(
		common.Commands,
		plugins.Commands,
	),
}
