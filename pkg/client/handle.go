package client

import (
	"errors"
	"fmt"

	"github.com/lucheng0127/vtun/pkg/iface"
	"github.com/lucheng0127/vtun/pkg/protocol"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

func (c *Client) SendPkt(pkt *protocol.VTPacket) error {
	stream, err := pkt.Encode()
	if err != nil {
		return err
	}

	_, err = c.Conn.Write(stream)
	return err
}

func (c *Client) SendReq() error {
	pkt, err := protocol.NewVTPkt(protocol.HDR_FLG_REQ, []byte(fmt.Sprintf("%s/%s", c.User, c.Passwd)), c.Cipher)
	if err != nil {
		return err
	}

	return c.SendPkt(pkt)
}

func (c *Client) HandleAck(payload []byte) error {
	msg := string(payload)
	// Parse msg, ip or error
	ipAddr, err := netlink.ParseAddr(msg)
	if err != nil {
		// Msg is error msg not ip address
		return errors.New(msg)
	}

	log.Infof("Connect to server succeed, endpoint ip %s", ipAddr.String())
	// Create tun
	iface, err := iface.SetupTun(ipAddr)
	if err != nil {
		return err
	}

	c.Iface = iface
	return nil
}

func (c *Client) HandleDat(payload []byte) {}

func (c *Client) HandleFin() {}
