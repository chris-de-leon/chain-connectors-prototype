package common

import (
	"github.com/urfave/cli/v3"
)

var Commands = []*cli.Command{
	version,
	clean,
}
