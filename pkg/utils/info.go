package utils

import (
	"net"
	"strings"
)

func ParseAllowedIPs(ctx string) ([]*net.IPNet, error) {
	infos := strings.Split(ctx, ",")
	allowedIPs := make([]*net.IPNet, 0)

	for _, info := range infos {
		_, ipNet, err := net.ParseCIDR(info)
		if err != nil {
			return nil, err
		}

		allowedIPs = append(allowedIPs, ipNet)
	}

	return allowedIPs, nil
}
