package command

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/logutils"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
)

const (
	// Boot2docker related constants.
	// These constants are from github.com/boot2docker/boot2docker-cli
	B2DSshKeyFile string = "id_boot2docker"
	B2DSshServer  string = "localhost:2022"
	B2DSshUser    string = "docker"

	// ClosingTime is time to wait until all server is closing
	ClosingTime = 1 * time.Second
)

var (
	DefaultLocalServer  = "localhost:8080"
	DefaultRemoteServer = "localhost:8080"
)

type ForwardCommand struct {
	Meta
}

func (c *ForwardCommand) Run(args []string) int {

	var logLevel string
	flags := flag.NewFlagSet("forward", flag.ContinueOnError)
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

	// Proxy forwarding is only for OS which need boot2docker
	// to run docker.
	if runtime.GOOS != "darwin" {
		c.Ui.Error("You don't need to run port forwarding")
		return 0
	}

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
		// Need some time to closing work...
		time.Sleep(ClosingTime)
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

// PortforwardServer
type PortForwardServer struct {
	Logger       *log.Logger
	LocalServer  string
	RemoteServer string
}

// Start starts server
func (s *PortForwardServer) Start() (chan struct{}, chan error, error) {
	// Setup ssh auth method from boot2docker ssh key file
	authMethod, err := B2DSshAuthMethod()
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to construct ssh auth method for boot2docker: %s", err)
	}

	cfg := &ssh.ClientConfig{
		User: B2DSshUser,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
	}

	// Establish connection with SSH server
	sshConn, err := ssh.Dial("tcp", B2DSshServer, cfg)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to establish connection with SSH server %s: %s", B2DSshServer, err)
	}
	s.Logger.Printf("[DEBUG] Establish connection with SSH server %s", B2DSshServer)

	// Start local server to forward traffic to remote server
	localListener, err := net.Listen("tcp", s.LocalServer)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to start local server %s: %s", s.LocalServer, err)
	}
	s.Logger.Printf("[INFO] Start local server")
	s.Logger.Printf("[INFO] Listening on %s (Ready to connection)", s.LocalServer)

	doneCh, errCh := make(chan struct{}), make(chan error)

	// Watch the doneCh and close server connection
	go func() {
		select {
		case <-doneCh:
			s.Logger.Println("[INFO] Stop server and close ssh connection")
			localListener.Close()
			sshConn.Close()
		}
	}()

	go func() {
		for {

			// Accept request and start local connection
			localConn, err := localListener.Accept()
			if err != nil {
				errCh <- fmt.Errorf("failed to accept request: %s", err)
			}
			s.Logger.Printf("[DEBUG] Accept request")

			// Establish connection with remote server via SSH connection
			sshRemoteConn, err := sshConn.Dial("tcp", s.RemoteServer)
			if err != nil {
				errCh <- fmt.Errorf(
					"failed to establish connection with remote server on boot2docker:\n"+"%s\n"+
						"This error happens when kubelet is not working. "+
						"Check it via `docker ps` command.", err)
			}
			s.Logger.Printf("[DEBUG] Establish connection with remote server %s", s.RemoteServer)

			doneRWCh := make(chan struct{})
			go func() {
				s.Logger.Printf("[DEBUG] Start data transfer from remote server to local")
				_, err := io.Copy(localConn, sshRemoteConn)
				if err != nil {
					s.Logger.Printf(
						"[ERROR] Failed to transfer from remote server to local: %s", err)
				}
				doneRWCh <- struct{}{}
			}()

			go func() {
				s.Logger.Printf("[DEBUG] Start data transfer from local server to remote")
				_, err := io.Copy(sshRemoteConn, localConn)
				if err != nil {
					s.Logger.Printf("[ERROR] Failed to transfer from local server to remote: %s", err)
				}
				doneRWCh <- struct{}{}
			}()

			<-doneRWCh
			localConn.Close()
			sshRemoteConn.Close()
			s.Logger.Printf("[DEBUG] Finish forwarding")
		}
	}()

	return doneCh, errCh, nil
}

// B2DSshAuthMethod return ssh auth method for boot2docker.
// It reads & parses ssh key file and constructs auth method.
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
