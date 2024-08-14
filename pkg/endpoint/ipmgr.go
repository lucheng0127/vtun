package endpoint

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/lucheng0127/vtun/pkg/utils"
	"github.com/vishvananda/netlink"
)

type IPMgr struct {
	IPRange string
	MaskLen int
	IPStart net.IP
	IPCount int
	UsedIP  []int
	MLock   sync.Mutex
}

func NewIPMgr(ipRange string, maskLen int) (*IPMgr, error) {
	mgr := &IPMgr{
		IPRange: ipRange,
		MaskLen: maskLen,
		UsedIP:  make([]int, 0),
		MLock:   sync.Mutex{},
	}

	if err := mgr.ParseIPRange(); err != nil {
		return nil, err
	}

	return mgr, nil
}

func (mgr *IPMgr) ParseIPRange() error {
	ipStart, ipCount := utils.ValidateDHCPRange(mgr.IPRange)
	if ipCount == 0 {
		return fmt.Errorf("invalidate dhcp range %s, <ip start>-<ip end>", mgr.IPRange)
	}

	mgr.IPStart = ipStart
	mgr.IPCount = ipCount
	return nil
}

func (mgr *IPMgr) IPIdxInPool(idx int) bool {
	for _, i := range mgr.UsedIP {
		if idx == i {
			return true
		}
	}

	return false
}

func (mgr *IPMgr) IPInPool(ip net.IP) bool {
	ipInt := binary.BigEndian.Uint32(ip)
	ipStartInt := binary.BigEndian.Uint32(mgr.IPStart)

	idx := ipInt - ipStartInt

	return mgr.IPIdxInPool(int(idx))
}

func (mgr *IPMgr) IPFromIdx(idx int) net.IP {
	ipStartInt := binary.BigEndian.Uint32(mgr.IPStart)
	ipInt := ipStartInt + uint32(idx)

	ipBytes := make([]byte, 4)

	binary.BigEndian.PutUint32(ipBytes, ipInt)

	return net.IP(ipBytes)
}

func (mgr *IPMgr) IdxFromIP(ip net.IP) int {
	ipStartInt := binary.BigEndian.Uint32(mgr.IPStart)
	ipInt := binary.BigEndian.Uint32(ip)

	return int(ipInt - ipStartInt)
}

func (mgr *IPMgr) IPToIPAddr(ip net.IP) (*netlink.Addr, error) {
	return netlink.ParseAddr(fmt.Sprintf("%s/%d", ip.String(), mgr.MaskLen))
}

func (mgr *IPMgr) IPForUser(username string) (*netlink.Addr, error) {
	idx := utils.IdxFromString(mgr.IPCount, username)
	idxEnd := idx - 1

	for idx < mgr.IPCount {
		if idx == idxEnd {
			// Checked the last idx
			if mgr.IPIdxInPool(idx) {
				break
			}

			mgr.MLock.Lock()
			mgr.UsedIP = append(mgr.UsedIP, idx)
			mgr.MLock.Unlock()
			return mgr.IPToIPAddr(mgr.IPFromIdx(idx))
		}

		if mgr.IPIdxInPool(idx) {
			idx += 1

			if idx == mgr.IPCount {
				// Check from zero
				idx = 0
			}

			continue
		}

		mgr.MLock.Lock()
		mgr.UsedIP = append(mgr.UsedIP, idx)
		mgr.MLock.Unlock()
		return mgr.IPToIPAddr(mgr.IPFromIdx(idx))
	}

	return nil, errors.New("run out of ip")
}

func (mgr *IPMgr) ReleaseIP(ipAddr *netlink.Addr) {
	ip := ipAddr.IPNet.IP.To4()
	idx := mgr.IdxFromIP(ip)

	targetPoolIdx := -1

	for i, ipIdx := range mgr.UsedIP {
		if ipIdx == idx {
			targetPoolIdx = i
			break
		}
	}

	if targetPoolIdx == -1 {
		return
	}

	mgr.MLock.Lock()
	mgr.UsedIP = append(mgr.UsedIP[:targetPoolIdx], mgr.UsedIP[targetPoolIdx+1:]...)
	mgr.MLock.Unlock()
}
