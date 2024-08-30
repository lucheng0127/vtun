package client

import (
	"errors"
	"os"

	"github.com/lucheng0127/vtun/pkg/iface"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func (c *Client) HandleAck(authChan chan string, payload []byte) error {
	msg := string(payload)
	// Parse msg, ip or error
	ipAddr, err := netlink.ParseAddr(msg)
	if err != nil {
		// Msg is error msg not ip address
		return errors.New(msg)
	}

	log.Infof("connect to server succeed, endpoint ip %s", ipAddr.String())
	authChan <- "OK"
	// Create tun
	iface, err := iface.SetupTun(ipAddr)
	if err != nil {
		return err
	}

	c.Iface = iface
	c.IPAddr = ipAddr

	// Start send heartbeat
	go c.SendHeartbeat()
	return nil
}

func (c *Client) HandlePsh() {
	log.Debug("heartbeat received")
	c.Beat <- "ping"
}

func (c *Client) HandleFin() {
	log.Info("FIN pkt received, exist")
	os.Exit(0)
}
