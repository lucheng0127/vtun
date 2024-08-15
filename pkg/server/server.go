package server

import (
	"fmt"
	"net"
	"os"

	"github.com/lucheng0127/vtun/pkg/auth"
	"github.com/lucheng0127/vtun/pkg/cipher"
	"github.com/lucheng0127/vtun/pkg/endpoint"
	"github.com/lucheng0127/vtun/pkg/protocol"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
)

type Svc interface {
	Launch() error
	Teardown()
	HandleSignal(chan os.Signal)
}

type Server struct {
	Iface   *water.Interface
	Port    int
	Key     string
	Conn    *net.UDPConn
	Cipher  cipher.Cipher
	AuthMgr *auth.BaseAuthMgr
	IPMgr   *endpoint.IPMgr
	EPMgr   *endpoint.EndpointMgr
}

func NewServer(iface *water.Interface, ipRange, userDB, key string, port, maskLen int) (Svc, error) {
	svc := new(Server)

	svc.Iface = iface
	svc.Port = port
	svc.Key = key
	svc.AuthMgr = &auth.BaseAuthMgr{DB: userDB}
	svc.EPMgr = endpoint.NewEPMgr()

	ipMgr, err := endpoint.NewIPMgr(ipRange, maskLen)
	if err != nil {
		return nil, err
	}

	svc.IPMgr = ipMgr
	return svc, nil
}

func (svc *Server) Launch() error {
	// Listen udp port
	addr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", svc.Port))
	if err != nil {
		return err
	}

	ln, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return err
	}
	svc.Conn = ln
	defer ln.Close()
	log.Infof("vtun server run on udp port %d", svc.Port)

	// Use aes cipher
	aesCipher, err := cipher.NewAESCipher(svc.Key)
	if err != nil {
		return err
	}
	svc.Cipher = aesCipher

	// Handle connection
	for {
		var buf [protocol.MAX_FRG_SIZE]byte
		n, raddr, err := ln.ReadFromUDP(buf[:])
		if err != nil {
			// Server not work well, return error
			return err
		}

		flag, payload, err := protocol.Decode(buf[:n], svc.Cipher)
		if err != nil {
			log.Warnf("parse VTPkt from address %s %s", raddr.String(), err)
			continue
		}

		switch flag {
		case protocol.HDR_FLG_REQ:
			svc.HandleReq(payload, raddr)
		case protocol.HDR_FLG_PSH:
			svc.HandlePsh(raddr)
		case protocol.HDR_FLG_DAT:
			svc.HandleDat(payload, raddr)
		case protocol.HDR_FLG_FIN:
			svc.HandleFin(raddr)
		default:
			// For server only hand req for user login
			// psh for keepalive
			// data for forward
			// fin for close endpoint
			continue
		}
	}
}

func (svc *Server) Teardown() {
	// TODO: Send FIN
	svc.Conn.Close()
}

func (svc *Server) HandleSignal(sigChan chan os.Signal) {
	sig := <-sigChan
	log.Infof("received signal: %v, stop server", sig)
	svc.Teardown()
	os.Exit(0)
}
