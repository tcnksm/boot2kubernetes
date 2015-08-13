package command

import (
	"bytes"
	"fmt"
	"time"

	"github.com/tcnksm/go-latest"
)

type VersionCommand struct {
	Meta

	Name     string
	Version  string
	Revision string
}

func (c *VersionCommand) Run(args []string) int {

	// Start goroutine to check latest version on Github
	verCheckCh := make(chan *latest.CheckResponse)
	go func() {
		githubTag := &latest.GithubTag{
			Owner:      "tcnksm",
			Repository: "boot2kubernetes",
		}

		// Ignore error, because it's not important
		res, _ := latest.Check(githubTag, c.Version)
		verCheckCh <- res
	}()

	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "%s version %s", c.Name, c.Version)
	if c.Revision != "" {
		fmt.Fprintf(&versionString, " (%s)", c.Revision)
	}

	c.Ui.Output(versionString.String())

	select {
	case res := <-verCheckCh:
		if res != nil && !res.Latest {
			c.Ui.Error(fmt.Sprintf(
				"Latest version of %s is %s, please update it", c.Name, res.Current))
		}
	case <-time.After(2 * time.Second):
		// Terminate version check soon
	}

	return 0
}

func (c *VersionCommand) Synopsis() string {
	return fmt.Sprintf("Print %s version and quit", c.Name)
}

func (c *VersionCommand) Help() string {
	return ""
}
