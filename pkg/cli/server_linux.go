package cli

import (
	"github.com/lucheng0127/vtun/pkg/server"
	"github.com/urfave/cli/v2"
)

// Server can only run on linux
func NewServerCmd() *cli.Command {
	return &cli.Command{
		Name:   "server",
		Usage:  "run vtun server",
		Action: server.Run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config-dir",
				Aliases:  []string{"d"},
				Usage:    "config directory to launch vtun server, config.yaml as config file, users as user storage",
				Required: true,
			},
		},
	}
}
