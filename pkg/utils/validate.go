package utils

import (
	"encoding/binary"
	"net"
	"strings"
)

func ValidateKey(key string) bool {
	switch len(key) {
	case 16, 24, 32:
		return true
	default:
		return false
	}
}

func ValidateDHCPRange(ipRange string) (net.IP, int) {
	ipRangeInfos := strings.Split(ipRange, "-")

	if len(ipRangeInfos) != 2 {
		return nil, 0
	}

	start := net.ParseIP(ipRangeInfos[0]).To4()
	end := net.ParseIP(ipRangeInfos[1]).To4()

	if start == nil || end == nil {
		return nil, 0
	}

	sInt := binary.BigEndian.Uint32(start)
	eInt := binary.BigEndian.Uint32(end)
	if eInt < sInt {
		return nil, 0
	}

	return start, int(eInt - sInt)
}
