package command

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/docker/libcompose/docker"
	"github.com/docker/libcompose/project"
	"github.com/hashicorp/logutils"
	"github.com/samalba/dockerclient"
	"github.com/tcnksm/boot2kubernetes/config"
)

const (
	// CheckInterval is how often check k8s container is ready
	CheckInterval = 3 * time.Second

	// CheckTimeout is timeout for waiting k8s container is ready
	CheckTimeOut = 300 * time.Second
)

type UpCommand struct {
	Meta
}

func (c *UpCommand) Run(args []string) int {
	var insecure bool
	var logLevel string
	flags := flag.NewFlagSet("up", flag.ContinueOnError)
	flags.BoolVar(&insecure, "insecure", false, "")
	flags.StringVar(&logLevel, "log-level", "info", "")
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

	// Setup new docker-compose project
	project, err := docker.NewProject(context)
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Failed to setup project: %s", err))
		return 1
	}

	c.Ui.Output("Start kubernetes cluster!")
	upErrCh := make(chan error)
	go func() {
		if err := project.Up(); err != nil {
			upErrCh <- err
		}
	}()

	client := clientFactory.Create(nil)

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case <-afterContainerReady(client):
		c.Ui.Info("Successfully start kubernetes cluster")
	case err := <-upErrCh:
		c.Ui.Error("")
		c.Ui.Error(fmt.Sprintf("Failed to start containers: %s", err))
		c.Ui.Error("Check docker daemon is wroking")
		return 1
	case <-sigCh:
		c.Ui.Error("")
		c.Ui.Error("Interrupted!")
		c.Ui.Error("It's ambiguous that boot2kubernetes could correctly start containers.")
		c.Ui.Error("So request to kubelet may be failed. Check the containers are working")
		c.Ui.Error("with `docker ps` command by yourself.")
		return 1
	case <-time.After(CheckTimeOut):
		c.Ui.Error("")
		c.Ui.Error("Timeout happened while waiting cluster containers are ready.")
		c.Ui.Error("It's ambiguous that boot2kubernetes could correctly start containers.")
		c.Ui.Error("So request to kubelet may be failed. Check the containers are working")
		c.Ui.Error("with `docker ps` command by yourself.")
		return 1
	}

	// If docker runs on boot2docker, port forwarding is needed.
	if runtime.GOOS == "darwin" {

		c.Ui.Output("")
		c.Ui.Output("==> WARNING: You're running docker on boot2docker!")
		c.Ui.Output("  To connect to master api server from local environment,")
		c.Ui.Output("  port forwarding is needed. boot2kubernetes starts ")
		c.Ui.Output("  server for that. To stop server, use ^C (Interrupt).\n")

		// Create logger with Log level
		logger := log.New(&logutils.LevelFilter{
			Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
			MinLevel: (logutils.LogLevel)(strings.ToUpper(logLevel)),
			Writer:   os.Stderr,
		}, "", log.LstdFlags)
		logger.Printf("[DEBUG] LogLevel: %s", logLevel)

		// Setup port forward server
		server := &PortForwardServer{
			Logger:       logger,
			LocalServer:  DefaultLocalServer,
			RemoteServer: DefaultRemoteServer,
		}

		doneCh, errCh, err := server.Start()
		if err != nil {
			c.Ui.Error(fmt.Sprintf(
				"Failed to start port forwarding server: %s", err))
			return 1
		}

		sigCh := make(chan os.Signal)
		signal.Notify(sigCh, os.Interrupt)
		select {
		case err := <-errCh:
			c.Ui.Error(fmt.Sprintf(
				"Error while running port forwarding server: %s", err))
			close(doneCh)
			return 1
		case <-sigCh:
			c.Ui.Error("\nInterrupted!")
			close(doneCh)
			// Need some time for closing work...
			time.Sleep(ClosingTime)
		}
	}

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

// afterContainerReady waits for the cluster ready and then sends the struct{}
// on the returned channel. Detection of cluster ready is very heuristic way,
// just checking number of container which is needed for running cluster.
func afterContainerReady(c dockerclient.Client) chan struct{} {
	doneCh := make(chan struct{})

	// Marshaling to post filter as API request
	filterLocalMasterStr, err := json.Marshal(FilterLocalMaster)
	if err != nil {
		// Should not reach here....
		panic(fmt.Sprintf(
			"Failed to marshal FilterLocalMaster: %s", err))
	}

	ticker := time.NewTicker(CheckInterval)
	go func() {
		fmt.Fprintf(os.Stderr, "Wait until containers are readly")
		for _ = range ticker.C {
			fmt.Fprintf(os.Stderr, ".")
			// Get Container info from deamon based on fileter
			localMasters, err := c.ListContainers(true, false, (string)(filterLocalMasterStr))
			if err != nil {
				// Just ignore error
				continue
			}

			if len(localMasters) > 3 {
				fmt.Fprintf(os.Stderr, "\n")
				doneCh <- struct{}{}
				ticker.Stop()
			}
		}
	}()

	return doneCh
}
