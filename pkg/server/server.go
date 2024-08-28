package server

import (
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/lucheng0127/vtun/pkg/auth"
	"github.com/lucheng0127/vtun/pkg/cipher"
	"github.com/lucheng0127/vtun/pkg/endpoint"
	"github.com/lucheng0127/vtun/pkg/iface"
	"github.com/lucheng0127/vtun/pkg/protocol"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water/waterutil"
	"github.com/vishvananda/netlink"
)

type Server struct {
	Iface   iface.IFace
	IPAddr  *netlink.Addr
	Port    int
	Key     string
	Conn    *net.UDPConn
	Cipher  cipher.Cipher
	AuthMgr *auth.BaseAuthMgr
	IPMgr   *endpoint.IPMgr
	EPMgr   *endpoint.EndpointMgr
	HbMgr   *HeartbeatMgr
	DstMgr  *DstMgr
	Routes  []*net.IPNet
}

func NewServer(ipRange, userDB, key string, port int, ipAddr *netlink.Addr, routes []string) (Svc, error) {
	svc := new(Server)
	svc.Routes = make([]*net.IPNet, 0)

	for _, route := range routes {
		_, rNet, err := net.ParseCIDR(route)
		if err != nil {
			return nil, err
		}

		svc.Routes = append(svc.Routes, rNet)
	}

	svc.IPAddr = ipAddr
	svc.Port = port
	svc.Key = key
	svc.AuthMgr = &auth.BaseAuthMgr{
		DB:         userDB,
		AuthedUser: map[string]string{},
		MLock:      sync.Mutex{},
	}
	svc.EPMgr = endpoint.NewEPMgr()
	svc.HbMgr = NewHeartbeatMgr(svc)

	maskLen, _ := ipAddr.IPNet.Mask.Size()
	ipMgr, err := endpoint.NewIPMgr(ipRange, maskLen)
	if err != nil {
		return nil, err
	}

	dstMgr := &DstMgr{
		InNet:    ipAddr.IPNet,
		ExNetMap: make(map[*net.IPNet]*endpoint.Endpoint),
		MLock:    sync.Mutex{},
		Svc:      svc,
	}

	svc.DstMgr = dstMgr
	svc.IPMgr = ipMgr
	return svc, nil
}

func (svc *Server) RouteAdd() error {
	return iface.RoutiesAdd(svc.Iface.Name(), svc.Routes, "")
}

func (svc *Server) PostUp() {
	// Add routes
	if err := svc.RouteAdd(); err != nil {
		log.Warnf("post up route add %s", err.Error())
	}
}

func (svc *Server) Launch() error {
	// Setup local tun
	iface, err := iface.SetupTun(svc.IPAddr)
	if err != nil {
		return err
	}
	svc.Iface = iface

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

	// Forward local tun traffic to net
	go svc.IfaceToNet()

	// PostUP
	go svc.PostUp()

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
			if err := svc.HandleReq(payload, raddr); err != nil {
				log.Error(err)
			}
		case protocol.HDR_FLG_PSH:
			svc.HandlePsh(raddr)
		case protocol.HDR_FLG_DAT:
			svc.HandleDat(payload)
		case protocol.HDR_FLG_FIN:
			svc.HandleFin(raddr)
		case protocol.HDR_FLG_IPS:
			svc.HandleIps(raddr.String(), payload)
		default:
			// For server only hand req for user login
			// psh for keepalive
			// data for forward
			// fin for close endpoint
			// ips for sync endpoint allowed-ips info
			continue
		}
	}
}

func (svc *Server) Teardown() {
	// Send FIN to all Endpoint
	log.Info("teardown server, send FIN pkt to all Endpoint")
	for ip := range svc.EPMgr.EPIPMap {
		svc.CloseEPByIP(ip)
	}

	svc.Conn.Close()
}

func (svc *Server) HandleSignal(sigChan chan os.Signal) {
	sig := <-sigChan
	log.Infof("received signal: %v, stop server", sig)
	svc.Teardown()
	os.Exit(0)
}

func (svc *Server) CloseEPByIP(ip string) {
	log.Infof("close Endpoint ip %s", ip)

	// Release ip and close Endpoint
	ep := svc.EPMgr.GetEPByIP(ip)
	if ep == nil {
		ipAddr, err := netlink.ParseAddr(ip)
		if err != nil {
			log.Errorf("close Endpoint by ip %s %s", ip, err.Error())
			return
		}

		svc.IPMgr.ReleaseIP(ipAddr)
		return
	}

	// Release allowed ip info with DstMgr
	svc.DstMgr.DelExNetByEp(ep)
	log.Infof("delete endpoint ip %s remote address %s allowed ip infos from dst mgr", ep.IP.String(), ep.RAddr.String())

	// Maybe client close, just send a FIN pkt
	log.Debugf("send FIN pkt to Endpoint remote address %s, ip %s", ep.RAddr.String(), ep.IP.String())
	if err := svc.SendFin(ep.RAddr); err != nil {
		log.Warn(err)
	}

	svc.IPMgr.ReleaseIP(ep.IP)
	svc.AuthMgr.LogoutUser(ep.User)

	if err := svc.EPMgr.CloseEPByIP(ip); err != nil {
		log.Error(err)
	}
}

func (svc *Server) GetDstEpByDstIP(dst net.IP) *endpoint.Endpoint {
	// Endpoint add allowed ip cidr, if dst to allowed ip, send to target endpoint
	return svc.DstMgr.GetDstEpByDstIP(dst)
}

func (svc *Server) IfaceToNet() {
	for {
		var buf [protocol.MAX_FRG_SIZE]byte

		n, err := svc.Iface.Read(buf[:])
		if err != nil {
			log.Error(err)
			continue
		}

		dst := waterutil.IPv4Destination(buf[:n])
		dstEp := svc.GetDstEpByDstIP(dst)

		if dstEp != nil {
			if err := svc.SendDat(buf[:n], dstEp.RAddr); err != nil {
				log.Errorf("forward traffic dst to %s to Endpoint %s %s", dst.String(), dstEp.RAddr.String(), err.Error())
				continue
			}

		}

		// If dstEp is nil, it means packet from local tun dst unknow
	}
}
