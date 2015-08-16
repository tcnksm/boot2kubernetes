package command

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	"github.com/samalba/dockerclient"
	"github.com/tcnksm/boot2kubernetes/config"
)

var FilterLocalMaster = map[string][]string{
	"label": []string{"io.kubernetes.pod.name=default/k8s-master-127.0.0.1"},
}

var FilterK8SRelated = map[string][]string{
	"label": []string{"io.kubernetes.pod.name"},
}

type DestroyCommand struct {
	Meta
}

func (c *DestroyCommand) Run(args []string) int {

	var insecure bool
	flags := flag.NewFlagSet("destroy", flag.ContinueOnError)
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

	// Setup new docker-compose project
	context := &docker.Context{
		Context: project.Context{
			Log:          false,
			ComposeBytes: compose,
			ProjectName:  "boot2k8s",
		},
		ClientFactory: clientFactory,
	}

	project, err := docker.NewProject(context)
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Failed to setup project: %s", err))
		return 1
	}

	if err := project.Delete(); err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Failed to destroy project: %s", err))
		return 1
	}

	client := clientFactory.Create(nil)

	// Marshaling to post filter as API request
	filterLocalMasterStr, err := json.Marshal(FilterLocalMaster)
	if err != nil {
		return 1
	}

	// Get Container info from deamon based on fileter
	localMasters, err := client.ListContainers(true, false, (string)(filterLocalMasterStr))
	if err != nil {
		return 1
	}

	if len(localMasters) > 0 {
		c.Ui.Output("Are you sure you want to destroy below containers?")
		for _, container := range localMasters {
			c.Ui.Output(fmt.Sprintf("  %s", container.Names[0]))
		}

		if yes, err := AskYesNo(); !yes || err != nil {
			if err == nil {
				c.Ui.Info("Containers will no be destroyed, since the confirmation")
				return 0
			}
			c.Ui.Error(fmt.Sprintf(
				"Terminate to destroy: %s", err.Error()))
			return 1
		}

		resultCh, errCh := removeContainers(client, localMasters, true, true)
		go func() {
			for res := range resultCh {
				c.Ui.Output(fmt.Sprintf(
					"Successfully destroy %s", res.Names[0]))
			}
		}()

		for err := range errCh {
			c.Ui.Error(fmt.Sprintf("Error: %s", err))
		}
		c.Ui.Output("")
	}

	// Marshaling to post filter as API request
	filterK8SRelatedStr, err := json.Marshal(FilterK8SRelated)
	if err != nil {
		return 1
	}

	relatedContainers, err := client.ListContainers(true, false, (string)(filterK8SRelatedStr))
	if err != nil {
		return 1
	}

	if len(relatedContainers) < 1 {
		// Correctly clean all containers
		return 0
	}

	c.Ui.Output("Do you also remove these containers? (these are created by kubernetes)")
	c.Ui.Error("==> WARNING: boot2kubernetes can not detect below containers")
	c.Ui.Error("  are created by kubernetes which up by boot2kubernetes.")
	c.Ui.Error("  Be sure below these will not be used anymore!")
	for _, container := range relatedContainers {
		c.Ui.Output(fmt.Sprintf("  %s", container.Names[0]))
	}

	if yes, err := AskYesNo(); !yes || err != nil {
		if err == nil {
			c.Ui.Info("Containers will no be destroyed, since the confirmation")
			return 0
		}
		c.Ui.Error(fmt.Sprintf(
			"Terminate to destroy: %s", err.Error()))
		return 1
	}

	resultCh, errCh := removeContainers(client, relatedContainers, true, true)
	go func() {
		for res := range resultCh {
			c.Ui.Output(fmt.Sprintf(
				"Successfully removed %s", res.Names[0]))
		}
	}()

	for err := range errCh {
		c.Ui.Error(fmt.Sprintf("Error: %s", err))
	}

	return 0
}

func (c *DestroyCommand) Synopsis() string {
	return "Destroy kubernetes cluster"
}

func (c *DestroyCommand) Help() string {
	helpText := `Destroy kubernetes cluseter.

Options:

  -insecure    Allow insecure non-TLS connection to docker client. 
`
	return strings.TrimSpace(helpText)
}

// removeContainers removes all containers parallelly.
// It retuns error channel and if something wrong, error is sent there.
func removeContainers(client dockerclient.Client, containers []dockerclient.Container, force, delVolume bool) (chan dockerclient.Container, chan error) {

	var wg sync.WaitGroup
	resultCh, errCh := make(chan dockerclient.Container), make(chan error)
	for _, container := range containers {
		wg.Add(1)
		go func(c dockerclient.Container) {
			defer wg.Done()
			if err := client.RemoveContainer(c.Id, force, delVolume); err != nil {
				errCh <- fmt.Errorf(
					"failed to remove %s (%s): %s", c.Names[0], c.Id, err)
				return
			}
			resultCh <- c
		}(container)
	}

	go func() {
		// Wait until all remove task and close error channnel then
		wg.Wait()
		close(resultCh)
		close(errCh)
	}()

	return resultCh, errCh
}

func AskYesNo() (bool, error) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	ansCh := make(chan bool, 1)
	go func() {
		for {
			fmt.Fprintf(os.Stderr, "Your choice? (Y/n) [default: n]: ")

			reader := bufio.NewReader(os.Stdin)
			line, _ := reader.ReadString('\n')
			line = strings.TrimRight(line, "\n")

			// Use Default value
			if line == "Y" {
				ansCh <- true
				break
			}

			if line == "n" || line == "" {
				ansCh <- false
				break
			}
		}
	}()

	select {
	case <-sigCh:
		return false, fmt.Errorf("interrupted")
	case yes := <-ansCh:
		return yes, nil
	}
}
