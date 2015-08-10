package command

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/logutils"
	"github.com/mitchellh/go-homedir"

	// This should be temporary, libcompose imports ssh package
	"github.com/tcnksm/boot2kubernetes/vendor/ssh"
)

const (
	B2DSshKeyFile string = "id_boot2docker"
	B2DSshUser    string = "docker"
)

type ForwardCommand struct {
	Meta
}

// B2DSshAuthMethod return ssh auth method for boot2docker.
// It reads & parse ssh key file and constuct auth method.
// If something wrong, returns error.
func B2DSshAuthMethod() (ssh.AuthMethod, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	// Read default file
	keyPath := filepath.Join(home, ".ssh", B2DSshKeyFile)
	buff, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	signer, _ := ssh.ParsePrivateKey(buff)
	if err != nil {
		return nil, err
	}

	return ssh.PublicKeys(signer), nil
}

func (c *ForwardCommand) Run(args []string) int {

	// Proxy forwarding is only for OS which need boot2docker
	// to run docker.
	if runtime.GOOS != "darwin" {
		c.Ui.Error("You don't need to run port forwarding")
		return 1
	}

	logLevel := "debug"

	// Create logger with Log level
	logger := log.New(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: (logutils.LogLevel)(strings.ToUpper(logLevel)),
		Writer:   os.Stderr,
	}, "", log.LstdFlags)

	logger.Printf("[DEBUG] LogLevel: %s", logLevel)

	authMethod, err := B2DSshAuthMethod()
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Failed to construct ssh auth method for boot2docker: %s", err))
		return 1
	}

	cfg := &ssh.ClientConfig{
		User: B2DSshUser,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
	}

	// Establish connection with SSH server
	sshServer := "localhost:2022"
	sshConn, err := ssh.Dial("tcp", sshServer, cfg)
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Failed to establish connection with SSH server %s: %s", sshServer, err))
		return 1
	}
	defer sshConn.Close()
	logger.Printf("[DEBUG] Establish connection with SSH server %s", sshServer)

	// Start local server to forward traffic to remote server
	localServer := "localhost:8080"
	localListener, err := net.Listen("tcp", localServer)
	if err != nil {
		c.Ui.Error(fmt.Sprintf(
			"Failed to start local server %s: %s", localServer, err))
		return 1
	}
	defer localListener.Close()
	logger.Printf("[INFO] Start local server")

	for {

		logger.Printf("[INFO] Listening on %s", localServer)
		// Accept requst and start local connection
		localConn, err := localListener.Accept()
		if err != nil {
			c.Ui.Error(fmt.Sprintf(
				"Failed to accept request: %s", err))
			return 1
		}
		logger.Printf("[INFO] Accept request")

		// Establish connection with remote server via SSH connection
		sshRemoteConn, err := sshConn.Dial("tcp", "localhost:8080")
		if err != nil {
			c.Ui.Error(fmt.Sprintf(
				"Failed to establish connection with remote server on boot2docker: %s\n"+
					"This error happends when kubelet is not working. Check it via `docker ps` command.", err))
			return 1
		}

		doneCh := make(chan struct{})

		go func() {
			logger.Printf("[DEBUG] Start data transfer from remote server to local")
			_, err := io.Copy(localConn, sshRemoteConn)
			if err != nil {
				logger.Printf("[ERROR] Failed to transfer from remote server to local: %s", err)
			}
			doneCh <- struct{}{}
		}()

		go func() {
			logger.Printf("[DEBUG] Start data transfer from local server to remote")
			_, err := io.Copy(sshRemoteConn, localConn)
			if err != nil {
				logger.Printf("[ERROR] Failed to transfer from local server to remote: %s", err)
			}
			doneCh <- struct{}{}
		}()

		<-doneCh

		// Close all connections
		localConn.Close()
		sshRemoteConn.Close()

		logger.Printf("[INFO] Finish forwarding")
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
