package client

import (
	"fmt"
	"strings"

	"github.com/lucheng0127/vtun/pkg/protocol"
	log "github.com/sirupsen/logrus"
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

func (c *Client) SendPsh() error {
	pkt, err := protocol.NewVTPkt(protocol.HDR_FLG_PSH, make([]byte, 0), c.Cipher)
	if err != nil {
		return err
	}

	return c.SendPkt(pkt)
}

func (c *Client) SendFin() error {
	pkt, err := protocol.NewVTPkt(protocol.HDR_FLG_FIN, make([]byte, 0), c.Cipher)
	if err != nil {
		return err
	}

	return c.SendPkt(pkt)
}

func (c *Client) SendDat(payload []byte) error {
	pkt, err := protocol.NewVTPkt(protocol.HDR_FLG_DAT, payload, c.Cipher)
	if err != nil {
		return err
	}

	return c.SendPkt(pkt)
}

func (c *Client) SendIps() error {
	if len(c.AllowedIPs) == 0 {
		return nil
	}

	allowedIPs := strings.Join(c.AllowedIPs, ",")
	pkt, err := protocol.NewVTPkt(protocol.HDR_FLG_IPS, []byte(allowedIPs), c.Cipher)
	if err != nil {
		return err
	}

	log.Debugf("send allowed-ips %s", allowedIPs)
	return c.SendPkt(pkt)
}
