package client

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/lucheng0127/vtun/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

type C interface {
	Launch() error
	Teardown()
	HandleSignal(chan os.Signal)
}

func Run(cCtx *cli.Context) error {
	// Parse config
	cfgFile := cCtx.String("config-file")
	cfg, err := config.LoadCleintConfigFile(cfgFile)
	if err != nil {
		return err
	}

	// Config log
	switch strings.ToUpper(cfg.LogLevel) {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	log.SetOutput(os.Stdout)

	client, err := NewClient(cfg.Target, cfg.Key, cfg.User, cfg.Passwd, cfg.AllowedIPs, cfg.Routes)
	if err != nil {
		return err
	}

	// Handle signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go client.HandleSignal(sigChan)

	return client.Launch()
}
