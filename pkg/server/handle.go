package server

import (
	"net"
	"strings"

	"github.com/lucheng0127/vtun/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water/waterutil"
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

	// TODO: Check user already login
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

func (svc *Server) HandleDat(payload []byte) {
	dst := waterutil.IPv4Destination(payload)
	// Forward traffic to dst Endpoint if know, nor send to local tun

	dstEp := svc.GetDstEpByDstIP(dst)
	if dstEp == nil {
		if _, err := svc.Iface.Write(payload); err != nil {
			log.Errorf("forward traffic dst to %s to tun %s %s", dst.String(), svc.Iface.Name(), err.Error())
			return
		}

		return
	}

	if err := svc.SendDat(payload, dstEp.RAddr); err != nil {
		log.Errorf("forward traffic dst to %s to Endpoint %s %s", dst.String(), dstEp.RAddr.String(), err.Error())
		return
	}
}

func (svc *Server) HandleFin(raddr *net.UDPAddr) {
	ep := svc.EPMgr.GetEPByAddr(raddr.String())
	if ep == nil {
		return
	}

	svc.HbMgr.StopMonitorEPByIP(ep.IP.String())
	svc.CloseEPByIP(ep.IP.String())
}

func (svc *Server) HandleIps(epAddr string, payload []byte) {
	ep := svc.EPMgr.GetEPByAddr(epAddr)
	if ep == nil {
		return
	}

	allowedIPs, err := utils.ParseAllowedIPs(string(payload))
	if err != nil {
		log.Errorf("invalidate allowed ip info %s", err.Error())
		return
	}

	log.Infof("add allowed ip %s for endpoint with ip %s remote address %s", string(payload), ep.IP.String(), ep.RAddr.String())
	svc.DstMgr.AddEpExNet(ep, allowedIPs...)
}
