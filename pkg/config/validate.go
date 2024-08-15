package config

import (
	"net"

	"github.com/go-playground/validator/v10"
	"github.com/vishvananda/netlink"
)

func ValidateCIDR(fl validator.FieldLevel) bool {
	if _, _, err := net.ParseCIDR(fl.Field().String()); err != nil {
		return false
	}

	return true
}

func ValidateKeyLength(fl validator.FieldLevel) bool {
	k := len([]byte(fl.Field().String()))

	switch k {
	default:
		return false
	case 16, 24, 32:
		return true
	}
}

func ValidateIPv4Addr(fl validator.FieldLevel) bool {
	addr, err := netlink.ParseAddr(fl.Field().String())
	if err != nil {
		return false
	}

	if addr.IP.To4() == nil {
		return false
	}

	return true
}
