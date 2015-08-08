package main

import (
	"github.com/mitchellh/cli"
	"github.com/tcnksm/boot2k8s/command"
)

func Commands(meta *command.Meta) map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"up": func() (cli.Command, error) {
			return &command.UpCommand{
				Meta: *meta,
			}, nil
		},
		"down": func() (cli.Command, error) {
			return &command.DownCommand{
				Meta: *meta,
			}, nil
		},

		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Meta:     *meta,
				Version:  Version,
				Revision: Revision,
				Name:     Name,
			}, nil
		},
	}
}
