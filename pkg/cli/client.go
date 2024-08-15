package cli

import (
	"github.com/lucheng0127/vtun/pkg/client"
	"github.com/urfave/cli/v2"
)

func NewClientCmd() *cli.Command {
	return &cli.Command{
		Name:   "client",
		Usage:  "connect to vtun server",
		Action: client.Run,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config-file",
				Aliases:  []string{"c"},
				Usage:    "config file of client",
				Required: true,
			},
		},
	}
}
