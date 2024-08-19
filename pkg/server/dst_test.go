package server

import (
	"net"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/lucheng0127/vtun/pkg/endpoint"
	mock_server "github.com/lucheng0127/vtun/pkg/mock/server"
	"github.com/vishvananda/netlink"
)

func TestDstMgr_GetDstEpByDstIP(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	fakeSvc := mock_server.NewMockSvc(ctrl)

	ipAddr, _ := netlink.ParseAddr("192.168.223.254/24")
	_, eNet1, _ := net.ParseCIDR("10.66.0.0/16")
	_, eNet2, _ := net.ParseCIDR("192.168.224.0/24")
	tEp := new(endpoint.Endpoint)
	tEp.User = "fake ep user"

	mgr := &DstMgr{
		InNet:    ipAddr.IPNet,
		ExNetMap: make(map[*net.IPNet]*endpoint.Endpoint),
		MLock:    sync.Mutex{},
		Svc:      fakeSvc,
	}

	mgr.AddEpExNet(tEp, eNet1, eNet2)
	ip1 := net.ParseIP("10.66.0.1").To4()
	ip2 := net.ParseIP("192.168.224.1").To4()
	ip3 := net.ParseIP("192.168.225.1").To4()

	ep1 := mgr.GetDstEpByDstIP(ip1)
	if tEp != ep1 {
		t.Errorf("DstMgr.GetDstEpByDstIP = %v, want %v", ep1, tEp)
	}
	ep2 := mgr.GetDstEpByDstIP(ip2)
	if tEp != ep2 {
		t.Errorf("DstMgr.GetDstEpByDstIP = %v, want %v", ep2, tEp)
	}
	ep3 := mgr.GetDstEpByDstIP(ip3)
	if ep3 != nil {
		t.Errorf("DstMgr.GetDstEpByDstIP = %v, want nil", ep2)
	}

	lenExNet := len(mgr.ExNetMap)
	if lenExNet != 2 {
		t.Errorf("DstMgr.ExNetMap length = %d, want 2", lenExNet)
	}

	mgr.DelExNetByEp(tEp)
	lenExNet = len(mgr.ExNetMap)
	if lenExNet != 0 {
		t.Errorf("After DstMgr.DelExNetByEp DstMgr.ExNetMap length = %d, want 0", lenExNet)
	}
}
