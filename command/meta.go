package command

import (
	"io/ioutil"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
)

func init() {
	// libcomopse depends on logrus and it generates its log.
	// So stop generating it from here.
	logrus.SetOutput(ioutil.Discard)
}

// Meta contain the meta-option that nearly all subcommand inherited.
type Meta struct {
	Ui cli.Ui
}
