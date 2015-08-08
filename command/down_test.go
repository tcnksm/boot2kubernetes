package command

import (
	"testing"

	"github.com/mitchellh/cli"
)

func TestDownCommand_implement(t *testing.T) {
	var _ cli.Command = &DownCommand{}
}
