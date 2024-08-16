package server

import (
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

func (svc *Server) HandleReq(payload []byte, raddr *net.UDPAddr) error {
	// Do Auth
	authInfos := strings.Split(string(payload), "/")
	if len(authInfos) != 2 {
		errMsg := "invalidate REQ msg"
		log.Errorf("%s remote address %s", errMsg, raddr.String())
		return svc.SendAck(errMsg, raddr)
	}

	// Login succeed add to EPMgr
	user := authInfos[0]
	passwd := authInfos[1]

	if err := svc.AuthMgr.ValidateUser(user, passwd); err != nil {
		log.Errorf("auth for remote addres %s %s", raddr.String(), err.Error())
		return svc.SendAck(err.Error(), raddr)
	}

	ipAddr, err := svc.IPMgr.IPForUser(user)
	if err != nil {
		log.Errorf("get ip for remote %s with user %s %s", raddr.String(), user, err.Error())
		return svc.SendAck(err.Error(), raddr)
	}

	log.Infof("remote %s login with user %s asign ip %s", raddr.String(), user, ipAddr.String())

	if err := svc.EPMgr.NewEP(svc.Conn, raddr, user, ipAddr); err != nil {
		log.Errorf("new endpoint entry %s for remote %s", err.Error(), raddr.String())
		return svc.SendAck(err.Error(), raddr)
	}

	// Monitor heartbeat for Endpoint
	svc.HbMgr.MonitorEPByIP(ipAddr.String())

	return svc.SendAck(ipAddr.String(), raddr)
}

func (svc *Server) HandlePsh(raddr *net.UDPAddr) {
	ep := svc.EPMgr.GetEPByAddr(raddr.String())
	if ep == nil {
		return
	}

	em := svc.HbMgr.GetEPMonitorByIP(ep.IP.String())

	log.Debugf("hearbeat received from %s ip %s", raddr.String(), ep.IP.String())
	em.Beat <- "ping"
}

func (svc *Server) HandleDat(payload []byte, raddr *net.UDPAddr) {}

func (svc *Server) HandleFin(raddr *net.UDPAddr) {
	ep := svc.EPMgr.GetEPByAddr(raddr.String())
	if ep == nil {
		return
	}

	svc.HbMgr.StopMonitorEPByIP(ep.IP.String())
	svc.CloseEPByIP(ep.IP.String())
}
