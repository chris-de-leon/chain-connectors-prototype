package constants

import (
	"fmt"
)

const (
	OWNER = "chris-de-leon"
	REPO  = "chain-connectors-prototype"
)

func VersionWithPrefix() string {
	return fmt.Sprintf("v%s", VersionWithoutPrefix())
}

func VersionWithoutPrefix() string {
	return "1.0.1"
}
