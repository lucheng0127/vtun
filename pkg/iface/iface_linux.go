package iface

import (
	"github.com/lucheng0127/vtun/pkg/utils"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

// Create a new tun interface assign a ipv4 address and set link up
// addr: <ipv4 addr>/<netmask>
func SetupTun(addr *netlink.Addr) (*water.Interface, error) {
	suffix := utils.RandStr(4)
	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = "tun-" + suffix

	iface, err := water.New(config)
	if err != nil {
		return nil, err
	}

	link, err := netlink.LinkByName(iface.Name())
	if err != nil {
		return nil, err
	}

	err = netlink.AddrAdd(link, addr)
	if err != nil {
		return nil, err
	}

	err = netlink.LinkSetUp(link)
	if err != nil {
		return nil, err
	}

	return iface, nil
}
