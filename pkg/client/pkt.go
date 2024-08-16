package client

import (
	"fmt"

	"github.com/lucheng0127/vtun/pkg/protocol"
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
