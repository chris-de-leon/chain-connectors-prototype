package core

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/constants"
	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/gh"
	"github.com/urfave/cli/v3"
)

const FileModeExecutable = 0755

var Repo = gh.NewRepository(constants.OWNER, constants.REPO)

func PrintResults(cmd *cli.Command, results []string) error {
	output, err := json.MarshalIndent(map[string][]string{"Result": results}, "", " ")
	if err != nil {
		return err
	} else {
		fmt.Fprintln(cmd.Root().Writer, string(output))
	}
	return nil
}

func CleanDir(c *cli.Command, dir string, force bool) error {
	response := "y"

	if !force {
		fmt.Fprintf(c.Root().Writer, "Remove '%s' (y/n): ", dir)
		if _, err := fmt.Scanf("%s", &response); err != nil {
			return err
		}
	}

	if response == "y" {
		if err := os.RemoveAll(dir); err != nil {
			return err
		} else {
			fmt.Fprintf(c.Root().Writer, "Removed '%s'\n", dir)
		}
	} else {
		fmt.Fprintf(c.Root().Writer, "Skipping '%s'\n", dir)
	}

	return nil
}

func ErrExit(err error) error {
	return cli.Exit(err, 1)
}
