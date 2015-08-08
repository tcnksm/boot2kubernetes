package command

import (
	"testing"

	"github.com/mitchellh/cli"
)

func TestUpCommand_implement(t *testing.T) {
	var _ cli.Command = &UpCommand{}
}
