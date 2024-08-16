package server

import (
	"net"

	"github.com/lucheng0127/vtun/pkg/protocol"
)

func (svc *Server) SendPkt(pkt *protocol.VTPacket, raddr *net.UDPAddr) error {
	stream, err := pkt.Encode()
	if err != nil {
		return err
	}

	_, err = svc.Conn.WriteToUDP(stream, raddr)
	return err
}

func (svc *Server) SendAck(msg string, raddr *net.UDPAddr) error {
	pkt, err := protocol.NewVTPkt(protocol.HDR_FLG_ACK, []byte(msg), svc.Cipher)
	if err != nil {
		return err
	}

	return svc.SendPkt(pkt, raddr)
}

func (svc *Server) SendFin(raddr *net.UDPAddr) error {
	pkt, err := protocol.NewVTPkt(protocol.HDR_FLG_FIN, make([]byte, 0), svc.Cipher)
	if err != nil {
		return err
	}

	return svc.SendPkt(pkt, raddr)
}

func (svc *Server) SendDat(payload []byte, raddr *net.UDPAddr) error {
	pkt, err := protocol.NewVTPkt(protocol.HDR_FLG_DAT, payload, svc.Cipher)
	if err != nil {
		return err
	}

	return svc.SendPkt(pkt, raddr)
}
