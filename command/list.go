package command

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/docker/libcompose/docker"
)

type ListCommand struct {
	Meta
}

func (c *ListCommand) Run(args []string) int {

	var insecure bool
	flags := flag.NewFlagSet("list", flag.ContinueOnError)
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

	// Set up docker client
	clientFactory, err := docker.NewDefaultClientFactory(
		docker.ClientOpts{
			TLS: !insecure,
		},
	)

	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Failed to construct Docker client: %s", err))
		return 1
	}

	client := clientFactory.Create(nil)

	// Marshaling to post filter as API request
	filterK8SRelatedStr, _ := json.Marshal(FilterK8SRelated)
	relatedContainers, err := client.ListContainers(true, false, (string)(filterK8SRelatedStr))
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Failed to list containers: %s", err))
		return 1
	}

	if len(relatedContainers) < 1 {
		c.Ui.Info("There are no containers which are labeled io.kubernetes.pod.name")
		return 0
	}

	c.Ui.Output("NAME")
	for _, container := range relatedContainers {
		c.Ui.Output(fmt.Sprintf("%s", container.Names[0]))
	}

	return 0
}

func (c *ListCommand) Synopsis() string {
	return "List all containers which are labeled `io.kubernetes.pod.name`"
}

func (c *ListCommand) Help() string {
	return "List all containers which are labeled `io.kubernetes.pod.name`"
}
