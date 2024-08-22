package iface

import (
	"fmt"
	"net"
	"strings"

	"github.com/lucheng0127/vtun/pkg/utils"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

// Create a new tun interface assign a ipv4 address and set link up
// addr: <ipv4 addr>/<netmask//
func SetupTun(addr *netlink.Addr) (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}
	iface, err := water.New(config)

	if err != nil {
		return nil, err
	}

	ifconfigMgr, err := utils.NewCmdMgr("ifconfig")
	if err != nil {
		return nil, err
	}

	// XXX: For macos set tun ip must set dst ip, but we can set sip same as dip
	// ref: https://stackoverflow.com/questions/17510101/how-do-i-set-an-ip-address-for-tun-interface-on-osx-without-destination-address
	subcmd := strings.Split(fmt.Sprintf("%s %s %s up", iface.Name(), addr.IP.String(), addr.IP.String()), " ")
	if _, err := ifconfigMgr.Execute(subcmd...); err != nil {
		return nil, err
	}

	routeconfigMgr, err := utils.NewCmdMgr("route")
	_, inNet, err := net.ParseCIDR(addr.IPNet.String())
	if err != nil {
		return nil, err
	}

	subcmd = strings.Split(fmt.Sprintf("add -net %s -interface %s", inNet.String(), iface.Name()), " ")
	if _, err := routeconfigMgr.Execute(subcmd...); err != nil {
		return nil, err
	}

	return iface, nil
}
