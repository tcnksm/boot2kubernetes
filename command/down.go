package command

import (
	"strings"
)

type DownCommand struct {
	Meta
}

func (c *DownCommand) Run(args []string) int {
	// Write your code here
	return 0
}

func (c *DownCommand) Synopsis() string {
	return ""
}

func (c *DownCommand) Help() string {
	helpText := `

`
	return strings.TrimSpace(helpText)
}
