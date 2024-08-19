package server

import (
	"fmt"
	"net"
	"sync"

	"github.com/lucheng0127/vtun/pkg/endpoint"
)

type DstMgr struct {
	InNet    *net.IPNet
	ExNetMap map[*net.IPNet]*endpoint.Endpoint
	MLock    sync.Mutex
	Svc      Svc
}

func (mgr *DstMgr) GetDstEpByDstIP(dst net.IP) *endpoint.Endpoint {
	if mgr.InNet.Contains(dst) {
		keySuffix, _ := mgr.InNet.Mask.Size()
		ipKey := fmt.Sprintf("%s/%d", dst.String(), keySuffix)
		return mgr.Svc.(*Server).EPMgr.GetEPByIP(ipKey)
	}

	for eNet, ep := range mgr.ExNetMap {
		if eNet.Contains(dst) {
			return ep
		}
	}

	return nil
}

func (mgr *DstMgr) DelExNetByEp(ep *endpoint.Endpoint) {
	mgr.MLock.Lock()
	defer mgr.MLock.Unlock()

	for eNet, tEp := range mgr.ExNetMap {
		if tEp == ep {
			delete(mgr.ExNetMap, eNet)
		}
	}
}

func (mgr *DstMgr) AddEpExNet(ep *endpoint.Endpoint, eNets ...*net.IPNet) {
	mgr.MLock.Lock()
	defer mgr.MLock.Unlock()
	for _, eNet := range eNets {
		mgr.ExNetMap[eNet] = ep
	}
}
