package util

import "net"

// GetIPVersion returns "v4" or "v6" based on the IP address.
// Returns "v4" as default when ip is empty.
func GetIPVersion(ip string) string {
	if ip == "" {
		return "v4"
	}

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "v4"
	}

	if parsedIP.To4() != nil {
		return "v4"
	}
	return "v6"
}
