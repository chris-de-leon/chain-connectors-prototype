package plgn

import (
	"fmt"
	"strings"
)

type PluginNotFoundError struct {
	Choices []string
	ID      string
}

func (e *PluginNotFoundError) Error() string {
	return fmt.Sprintf(
		"plugin with ID '%s' does not exist - must be one of: [ %s ]",
		e.ID,
		strings.Join(e.Choices, ", "),
	)
}

func (e *PluginNotFoundError) Is(target error) bool {
	_, ok := target.(*PluginNotFoundError)
	return ok
}
