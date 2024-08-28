package iface

import (
	"net"

	"github.com/lucheng0127/vtun/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/songgao/water"
	"github.com/vishvananda/netlink"
)

type IFace interface {
	Close() error
	Name() string
	Read([]byte) (int, error)
	Write([]byte) (int, error)
}

func RoutiesAdd(iface string, routes []*net.IPNet, ip string) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return err
	}

	for _, rNet := range routes {
		log.Debugf("add route %s dev %s", rNet.String(), iface)

		route := &netlink.Route{
			Dst:       rNet,
			LinkIndex: link.Attrs().Index,
		}

		if err := netlink.RouteReplace(route); err != nil {
			return err
		}
	}

	return nil
}

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
