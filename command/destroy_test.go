package command

import (
	"testing"

	"github.com/mitchellh/cli"
)

func TestDestroyCommand_implement(t *testing.T) {
	var _ cli.Command = &DestroyCommand{}
}
