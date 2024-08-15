package endpoint

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/lucheng0127/vtun/pkg/protocol"
	"github.com/vishvananda/netlink"
)

// An Endpoint represent a authed client
type Endpoint struct {
	Conn      *net.UDPConn
	RAddr     *net.UDPAddr
	User      string
	IP        *netlink.Addr
	LoginTime string
	TXQueue   chan *protocol.VTPacket
}

func (ep *Endpoint) Close() error {
	// TODO: Implement it if need
	return nil
}

type EndpointMgr struct {
	MLock     sync.Mutex
	EPIPMap   map[string]string
	EPAddrMap map[string]*Endpoint
}

func NewEPMgr() *EndpointMgr {
	return &EndpointMgr{
		MLock:     sync.Mutex{},
		EPIPMap:   make(map[string]string),
		EPAddrMap: make(map[string]*Endpoint),
	}
}

func (mgr *EndpointMgr) GetEPAddrByIP(ip string) string {
	addr, ok := mgr.EPIPMap[ip]
	if ok {
		return addr
	}

	return ""
}

func (mgr *EndpointMgr) GetEPByAddr(addr string) *Endpoint {
	ep, ok := mgr.EPAddrMap[addr]
	if ok {
		return ep
	}

	return nil
}

func (mgr *EndpointMgr) NewEP(conn *net.UDPConn, raddr *net.UDPAddr, user string, ip *netlink.Addr) error {
	ep := &Endpoint{
		Conn:      conn,
		RAddr:     raddr,
		User:      user,
		IP:        ip,
		LoginTime: time.Now().Format("2006-01-02 15:04:05"),
		TXQueue:   make(chan *protocol.VTPacket),
	}
	addrKey := raddr.String()
	ipKey := ip.String()

	// Check ip or raddr exist
	if addr := mgr.GetEPAddrByIP(ipKey); addr != "" {
		return fmt.Errorf("IP address %s used by Endpoint with remote address %s", ipKey, addr)
	}

	if addrEP := mgr.GetEPByAddr(addrKey); addrEP != nil {
		return fmt.Errorf("Endpoint with remote address %s exist, ip %s", addrKey, addrEP.IP.String())
	}

	mgr.MLock.Lock()
	mgr.EPAddrMap[addrKey] = ep
	mgr.EPIPMap[ipKey] = addrKey
	mgr.MLock.Unlock()

	// TODO: Start countdown

	return nil
}

func (mgr *EndpointMgr) CloseEPByIP(ip string) error {
	mgr.MLock.Lock()
	defer mgr.MLock.Unlock()

	addrKey, ok := mgr.EPIPMap[ip]
	if !ok {
		return nil
	}

	ep, ok := mgr.EPAddrMap[addrKey]
	if !ok {
		delete(mgr.EPIPMap, ip)
		return nil
	}

	delete(mgr.EPAddrMap, addrKey)
	delete(mgr.EPIPMap, ip)
	if err := ep.Close(); err != nil {
		return err
	}

	return nil
}
