package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chris-de-leon/chain-connectors/src/cli/libs/cmd"
	"github.com/urfave/cli/v3"
)

func main() {
	if err := cmd.Commands.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(cli.ErrWriter, err)
	}
}
