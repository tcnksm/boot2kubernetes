package main

import (
	"github.com/mitchellh/cli"
	"github.com/tcnksm/boot2kubernetes/command"
)

func Commands(meta *command.Meta) map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"up": func() (cli.Command, error) {
			return &command.UpCommand{
				Meta: *meta,
			}, nil
		},
		"forward": func() (cli.Command, error) {
			return &command.ForwardCommand{
				Meta: *meta,
			}, nil
		},

		"destroy": func() (cli.Command, error) {
			return &command.DestroyCommand{
				Meta: *meta,
			}, nil
		},

		"list": func() (cli.Command, error) {
			return &command.ListCommand{
				Meta: *meta,
			}, nil
		},

		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Meta:     *meta,
				Version:  Version,
				Revision: GitCommit,
				Name:     Name,
			}, nil
		},
	}
}
