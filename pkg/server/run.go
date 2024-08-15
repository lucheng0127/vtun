package server

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/lucheng0127/vtun/pkg/config"
	"github.com/lucheng0127/vtun/pkg/iface"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"github.com/vishvananda/netlink"
)

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
	maskLen, _ := ipAddr.IPNet.Mask.Size()

	// Setup local tun
	iface, err := iface.SetupTun(ipAddr)
	if err != nil {
		return err
	}

	// Create server
	svc, err := NewServer(iface, cfg.IPRange, userDB, cfg.Port, maskLen)
	if err != nil {
		return err
	}

	// Handle signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go svc.HandleSignal(sigChan)

	// Launch
	return svc.Launch()
}
