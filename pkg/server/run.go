package server

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/lucheng0127/vtun/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/vishvananda/netlink"
)

type Svc interface {
	Launch() error
	Teardown()
	HandleSignal(chan os.Signal)
}

var VSvc Svc

func Run(cCtx *cli.Context) error {
	// Parse config
	cfgDir := cCtx.String("config-dir")
	cfgFile := config.GetCfgPath(cfgDir)
	userDB := config.GetUserDBPath(cfgDir)
	cfg, err := config.LoadServerConfigFile(cfgFile)
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

	// Config check
	ipAddr, err := netlink.ParseAddr(cfg.IP)
	if err != nil {
		return err
	}

	// Create server
	svc, err := NewServer(cfg.IPRange, userDB, cfg.Key, cfg.Port, ipAddr, cfg.Routes)
	if err != nil {
		return err
	}
	VSvc = svc

	// Handle signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go svc.HandleSignal(sigChan)

	// Launch web server
	if cfg.WebConfig != nil && cfg.WebConfig.Enable {
		webSvc := &WebServer{
			Port: cfg.WebConfig.Port,
		}

		go webSvc.Serve()
		log.Infof("run web server on port %d", webSvc.Port)
	}

	// Launch
	return svc.Launch()
}
