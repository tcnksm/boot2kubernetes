package command

import (
	"io"
	"io/ioutil"
	"log"
	"net"
	"path/filepath"
	"strings"

	// This should be temporary, libcompose imports ssh package
	"github.com/mitchellh/go-homedir"
	"github.com/tcnksm/boot2k8s/vendor/ssh"
)

const (
	Boot2DockerBin string = "boot2docker"
)

type ForwardCommand struct {
	Meta
}

func (c *ForwardCommand) Run(args []string) int {

	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	keyPath := filepath.Join(home, ".ssh", "id_boot2docker")
	buff, err := ioutil.ReadFile(keyPath)
	if err != nil {
		panic(err)
	}

	key, _ := ssh.ParsePrivateKey(buff)
	if err != nil {
		panic(err)
	}

	cfg := &ssh.ClientConfig{
		User: "docker",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	// Establish connection with SSH server
	sshConn, err := ssh.Dial("tcp", "localhost:2022", cfg)
	if err != nil {
		panic(err)
	}
	defer sshConn.Close()

	// Start local server to forward traffic to remote server
	localListener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}
	defer localListener.Close()

	for {

		// Establish connection with remote server via SSH connection
		sshRemoteConn, err := sshConn.Dial("tcp", "localhost:8080")
		if err != nil {
			panic(err)
		}

		// Accept requst and start local connection
		localConn, err := localListener.Accept()
		if err != nil {
			panic(err)
		}

		log.Printf("Accepted connection %v\n", localConn)

		doneCh := make(chan struct{})

		// Start data transfer from remote to local
		go func() {
			_, err := io.Copy(localConn, sshRemoteConn)
			if err != nil {
				panic(err)
			}
			doneCh <- struct{}{}
		}()

		// Start data transfer from local to remote
		go func() {
			_, err := io.Copy(sshRemoteConn, localConn)
			if err != nil {
				panic(err)
			}
			doneCh <- struct{}{}
		}()

		<-doneCh

		localConn.Close()
		sshRemoteConn.Close()
	}

	return 0
}

func (c *ForwardCommand) Synopsis() string {
	return "Run port forwarding server"
}

func (c *ForwardCommand) Help() string {
	helpText := `Run port forwarding server. 
`
	return strings.TrimSpace(helpText)
}
