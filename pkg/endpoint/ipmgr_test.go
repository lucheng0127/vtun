package endpoint

import (
	"net"
	"reflect"
	"sync"
	"testing"

	"bou.ke/monkey"
	"github.com/lucheng0127/vtun/pkg/utils"
	"github.com/vishvananda/netlink"
)

func TestIPMgr_IPForUser(t *testing.T) {
	addrIdx1, _ := netlink.ParseAddr("192.168.123.101/24")
	addrIdx2, _ := netlink.ParseAddr("192.168.123.102/24")
	addrIdx3, _ := netlink.ParseAddr("192.168.123.100/24")
	type args struct {
		username string
	}
	tests := []struct {
		name       string
		args       args
		want       *netlink.Addr
		wantErr    bool
		patchFunc  interface{}
		targetFunc interface{}
	}{
		{
			name:       "idx 1-1",
			args:       args{username: "whocares"},
			want:       addrIdx1,
			wantErr:    false,
			patchFunc:  utils.IdxFromString,
			targetFunc: func(int, string) int { return 1 },
		},
		{
			name:       "idx 1-2",
			args:       args{username: "whocares"},
			want:       addrIdx2,
			wantErr:    false,
			patchFunc:  utils.IdxFromString,
			targetFunc: func(int, string) int { return 1 },
		},
		{
			name:       "idx 1-3",
			args:       args{username: "whocares"},
			want:       addrIdx3,
			wantErr:    false,
			patchFunc:  utils.IdxFromString,
			targetFunc: func(int, string) int { return 1 },
		},
		{
			name:       "idx 1-4",
			args:       args{username: "whocares"},
			want:       nil,
			wantErr:    true,
			patchFunc:  utils.IdxFromString,
			targetFunc: func(int, string) int { return 1 },
		},
	}

	ipStart := net.ParseIP("192.168.123.100").To4()
	svc := &IPMgr{
		UsedIP:  make([]int, 0),
		IPStart: ipStart,
		IPCount: 3,
		MaskLen: 24,
		MLock:   sync.Mutex{},
	}

	for _, tt := range tests {
		if tt.targetFunc != nil {
			monkey.Patch(tt.patchFunc, tt.targetFunc)
		}

		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.IPForUser(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.IPForUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.IPForUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPMgr_ReleaseIP(t *testing.T) {
	mgr, err := NewIPMgr("192.168.123.100-192.168.123.200", 24)
	if err != nil {
		t.Errorf("NewIPMgr %v", err)
		return
	}

	// Padding used ip
	mgr.UsedIP = append(mgr.UsedIP, 0, 1)

	ip1, _ := netlink.ParseAddr("192.168.123.100/24")
	ip2, _ := netlink.ParseAddr("192.168.123.101/24")
	ip3, _ := netlink.ParseAddr("192.168.123.105/24")

	mgr.ReleaseIP(ip1)
	if len(mgr.UsedIP) != 1 {
		t.Errorf("ReleaseIP length of used ip want = 1 got = %d", len(mgr.UsedIP))
		return
	}

	mgr.ReleaseIP(ip1)
	if len(mgr.UsedIP) != 1 {
		t.Errorf("ReleaseIP length of used ip want = 1 got = %d", len(mgr.UsedIP))
		return
	}

	mgr.ReleaseIP(ip3)
	if len(mgr.UsedIP) != 1 {
		t.Errorf("ReleaseIP length of used ip want = 1 got = %d", len(mgr.UsedIP))
		return
	}

	mgr.ReleaseIP(ip2)
	if len(mgr.UsedIP) != 0 {
		t.Errorf("ReleaseIP length of used ip want = 0 got = %d", len(mgr.UsedIP))
		return
	}
}
