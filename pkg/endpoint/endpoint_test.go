package endpoint

import (
	"net"
	"testing"
	"time"

	"github.com/vishvananda/netlink"
)

func TestEndpointMgr_NewEP(t *testing.T) {
	// Old ep map infos
	ep1IP, _ := netlink.ParseAddr("192.168.0.1/24")
	ep1Addr, _ := net.ResolveUDPAddr("udp", "172.16.18.1:1234")

	ep2IP, _ := netlink.ParseAddr("192.168.0.2/24")
	ep2Addr, _ := net.ResolveUDPAddr("udp", "172.16.18.2:1234")

	mgr := NewEPMgr()
	mgr.EPIPMap[ep1IP.String()] = ep1Addr.String()
	mgr.EPAddrMap[ep1Addr.String()] = &Endpoint{
		RAddr:     new(net.UDPAddr),
		User:      "ep1User",
		IP:        ep1IP,
		LoginTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	mgr.EPAddrMap[ep2Addr.String()] = &Endpoint{
		RAddr:     new(net.UDPAddr),
		User:      "ep2User",
		IP:        ep2IP,
		LoginTime: time.Now().Format("2006-01-02 15:04:05"),
	}

	// Ep waiting for add
	weConn := new(net.UDPConn)
	weAddr, _ := net.ResolveUDPAddr("udp", "172.16.17.254:1234")
	weUser := "weUser"
	weIP, _ := netlink.ParseAddr("192.168.0.254/24")

	type args struct {
		conn  *net.UDPConn
		raddr *net.UDPAddr
		user  string
		ip    *netlink.Addr
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			args: args{
				conn:  weConn,
				raddr: weAddr,
				user:  weUser,
				ip:    weIP,
			},
			wantErr: false,
		},
		{
			name: "ip used",
			args: args{
				conn:  new(net.UDPConn),
				raddr: new(net.UDPAddr),
				user:  "ep1User",
				ip:    ep1IP,
			},
			wantErr: true,
		},
		{
			name: "addr used",
			args: args{
				conn:  new(net.UDPConn),
				raddr: ep2Addr,
				user:  "ep2User",
				ip:    ep2IP,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := mgr.NewEP(tt.args.conn, tt.args.raddr, tt.args.user, tt.args.ip); (err != nil) != tt.wantErr {
				t.Errorf("EndpointMgr.NewEP() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.name == "ok" {
				addr := mgr.GetEPAddrByIP(tt.args.ip.String())
				if addr != tt.args.raddr.String() {
					t.Errorf("GetEPAddrByIP() want = %v, got = %v", tt.args.raddr.String(), addr)
				}

				ep := mgr.GetEPByAddr(addr)
				if ep.IP.String() != tt.args.ip.String() {
					t.Errorf("Endpoint get by GetEPByAddr info not match, wantIP = %v, gotIP = %v", ep.IP.String(), tt.args.ip.String())
				}
			}
		})
	}
}
