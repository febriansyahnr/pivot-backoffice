package util

import (
	"net"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

func ValidateIPAddress(ip, cidr string) error {
	input := ip
	if cidr != "" {
		input = ip + "/" + cidr
		_, _, err := net.ParseCIDR(input) // return ip, network, error
		if err != nil {
			return err
		}
	} else {
		data := net.ParseIP(input)
		if data == nil {
			return constant.ErrInvalidIPAddress
		}
	}

	return nil
}

func IsIPIdentical(ip1, ip2 string) bool {
	firstIP := net.ParseIP(ip1)
	if firstIP == nil {
		return false
	}
	secondIP := net.ParseIP(ip2)
	if secondIP == nil {
		return false
	}
	return firstIP.Equal(secondIP)
}

func IsInIPRange(ipRange, ip string) bool {
	_, block, err := net.ParseCIDR(ipRange)
	if err != nil {
		return false
	}
	return block.Contains(net.ParseIP(ip))
}

func IsIPMatch(configIP, configSubnet string, requestIP string) bool {
	// Handle default IP for both IPv4 and IPv6
	if configIP == constant.DefaultIP || configIP == constant.DefaultIPv6 {
		return true
	}

	if configSubnet == "" {
		return IsIPIdentical(configIP, requestIP)
	} else {
		return IsInIPRange(configIP+"/"+configSubnet, requestIP)
	}
}
