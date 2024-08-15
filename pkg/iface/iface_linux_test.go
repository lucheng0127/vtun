package iface

import (
	"errors"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

type PatchObj struct {
	PatchFunc  interface{}
	TargetFunc interface{}
}

func TestSetupTun(t *testing.T) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = "tun-UseForTest"
	fakeIface, err := water.New(config)
	if err != nil {
		t.Errorf("Setup testcase %s", err)
		return
	}

	addr, err := netlink.ParseAddr("192.168.1.1/24")
	if err != nil {
		t.Errorf("Parse addr %s", err)
		return
	}

	type args struct {
		addr *netlink.Addr
	}
	tests := []struct {
		name      string
		args      args
		want      *water.Interface
		wantErr   bool
		patchList []*PatchObj
	}{
		{
			name:    "ok",
			args:    args{addr: addr},
			want:    fakeIface,
			wantErr: false,
			patchList: []*PatchObj{
				{
					PatchFunc: water.New,
					TargetFunc: func(water.Config) (*water.Interface, error) {
						return fakeIface, nil
					},
				},
			},
		},
		{
			name:    "set link up failed",
			args:    args{addr: addr},
			want:    nil,
			wantErr: true,
			patchList: []*PatchObj{
				{
					PatchFunc: netlink.LinkSetUp,
					TargetFunc: func(netlink.Link) error {
						return errors.New("set link up failed")
					},
				},
			},
		},
		{
			name:    "add ipv4 addr failed",
			args:    args{addr: addr},
			want:    nil,
			wantErr: true,
			patchList: []*PatchObj{
				{
					PatchFunc: netlink.AddrAdd,
					TargetFunc: func(netlink.Link, *netlink.Addr) error {
						return errors.New("add address failed")
					},
				},
			},
		},
		{
			name:    "tun not found",
			args:    args{addr: addr},
			want:    nil,
			wantErr: true,
			patchList: []*PatchObj{
				{
					PatchFunc: netlink.LinkByName,
					TargetFunc: func(string) (netlink.Link, error) {
						return nil, errors.New("tun not found")
					},
				},
			},
		},
		{
			name:    "create tun error",
			args:    args{addr: addr},
			want:    nil,
			wantErr: true,
			patchList: []*PatchObj{
				{
					PatchFunc: water.New,
					TargetFunc: func(water.Config) (*water.Interface, error) {
						return nil, errors.New("create tun failed")
					},
				},
			},
		},
	}
	for _, tt := range tests {
		// Monkey patch
		for _, obj := range tt.patchList {
			monkey.Patch(obj.PatchFunc, obj.TargetFunc)
		}

		t.Run(tt.name, func(t *testing.T) {
			got, err := SetupTun(tt.args.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetupTun() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetupTun() = %v, want %v", got, tt.want)
			}
		})
	}
}
