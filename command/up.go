package command

import (
	"strings"
)

type UpCommand struct {
	Meta
}

func (c *UpCommand) Run(args []string) int {
	// Write your code here
	return 0
}

func (c *UpCommand) Synopsis() string {
	return ""
}

func (c *UpCommand) Help() string {
	helpText := `

`
	return strings.TrimSpace(helpText)
}
