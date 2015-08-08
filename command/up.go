package command

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	"github.com/tcnksm/boot2k8s/config"
)

type UpCommand struct {
	Meta
}

func (c *UpCommand) Run(args []string) int {

	var insecure bool
	flags := flag.NewFlagSet("up", flag.ContinueOnError)
	flags.BoolVar(&insecure, "insecure", false, "")
	flags.Usage = func() { c.Ui.Error(c.Help()) }

	errR, errW := io.Pipe()
	errScanner := bufio.NewScanner(errR)
	go func() {
		for errScanner.Scan() {
			c.Ui.Error(errScanner.Text())
		}
	}()

	flags.SetOutput(errW)

	if err := flags.Parse(args); err != nil {
		return 1
	}

	compose, err := config.Asset("k8s.yml")
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Failed to read k8s.yml: %s", err))
		return 1
	}

	// Setup new docker-compose project
	project, err := docker.NewProject(&docker.Context{
		Context: project.Context{
			ComposeBytes: compose,
			ProjectName:  "boot2k8s",
		},
		Tls: !insecure,
	})

	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Failed to setup project: %s", err))
		return 1
	}

	if err := project.Up(); err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Failed to up project: %s", err))
		return 1
	}

	// Write your code here
	return 0
}

func (c *UpCommand) Synopsis() string {
	return "Up kubernetes cluster"
}

func (c *UpCommand) Help() string {
	helpText := `Up kubernetes cluseter

Options:

  -insecure    Allow insecure non-TLS connection to docker client. 
`
	return strings.TrimSpace(helpText)
}
