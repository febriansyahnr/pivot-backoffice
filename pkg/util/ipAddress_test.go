package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	ipAddress  = "237.84.2.178"
	ipAddress2 = "237.84.2.1781"
	ipAddress3 = "10.0.0.0"
	configIp   = "0.0.0.0"
	configIp2  = "10.0.0.0"
	configIp3  = "10.0.1.1"
	subnet     = "24"
	subnet2    = "-44"
	ipAddress4 = "10.0.0.0/24"
	ipAddress5 = "10.0.1.1/24"
	
	// IPv6 test constants
	ipv6Address1  = "2001:db8::1"
	ipv6Address2  = "2001:db8::2"
	ipv6Address3  = "2001:db8:1::1"
	ipv6Address4  = "fe80::1"
	configIPv6_1  = "::"
	configIPv6_2  = "2001:db8::"
	configIPv6_3  = "2001:db8::1"
	ipv6Subnet    = "64"
	ipv6Subnet2   = "128"
	ipv6Range1    = "2001:db8::/64"
	ipv6Range2    = "2001:db8:1::/64"
)

func TestValidateIPAddress(t *testing.T) {
	testCases := []struct {
		name    string
		ip      string
		cidr    string
		wantErr bool
	}{
		// IPv4 Valid Cases
		{
			name: "SUCCESS: Valid IPv4",
			ip:   ipAddress,
		},
		{
			name: "SUCCESS: Valid IPv4 Subnet",
			ip:   ipAddress,
			cidr: subnet,
		},
		{
			name: "SUCCESS: IPv4 localhost",
			ip:   "127.0.0.1",
		},
		{
			name: "SUCCESS: IPv4 private range",
			ip:   "192.168.1.1",
		},
		{
			name: "SUCCESS: IPv4 with /32 subnet",
			ip:   "192.168.1.1",
			cidr: "32",
		},
		{
			name: "SUCCESS: IPv4 with /0 subnet (any)",
			ip:   "1.1.1.1",
			cidr: "0",
		},
		
		// IPv6 Valid Cases
		{
			name: "SUCCESS: Valid IPv6",
			ip:   ipv6Address1,
		},
		{
			name: "SUCCESS: Valid IPv6 Subnet",
			ip:   ipv6Address1,
			cidr: ipv6Subnet,
		},
		{
			name: "SUCCESS: Valid IPv6 with /128",
			ip:   ipv6Address1,
			cidr: ipv6Subnet2,
		},
		{
			name: "SUCCESS: IPv6 localhost",
			ip:   "::1",
		},
		{
			name: "SUCCESS: IPv6 any address",
			ip:   "::",
		},
		{
			name: "SUCCESS: IPv6 link-local",
			ip:   "fe80::1",
		},
		{
			name: "SUCCESS: IPv6 with /0 subnet",
			ip:   "2001:db8::1",
			cidr: "0",
		},
		{
			name: "SUCCESS: IPv6 compressed notation",
			ip:   "2001:db8::8a2e:370:7334",
		},
		{
			name: "SUCCESS: IPv4-mapped IPv6",
			ip:   "::ffff:192.0.2.1",
		},
		
		// Invalid IP Address Cases
		{
			name:    "ERROR: Invalid IPv4 octet too large",
			ip:      ipAddress2,
			wantErr: true,
		},
		{
			name:    "ERROR: IPv4 empty string",
			ip:      "",
			wantErr: true,
		},
		{
			name:    "ERROR: IPv4 too many octets",
			ip:      "192.168.1.1.1",
			wantErr: true,
		},
		{
			name:    "ERROR: IPv4 too few octets",
			ip:      "192.168.1",
			wantErr: true,
		},
		{
			name:    "ERROR: IPv4 non-numeric",
			ip:      "192.168.a.1",
			wantErr: true,
		},
		{
			name:    "ERROR: IPv4 with spaces",
			ip:      "192.168. 1.1",
			wantErr: true,
		},
		{
			name:    "ERROR: IPv6 too many groups",
			ip:      "2001:db8:0:0:1:0:0:1:extra",
			wantErr: true,
		},
		{
			name:    "ERROR: IPv6 invalid hex",
			ip:      "2001:db8::gggg",
			wantErr: true,
		},
		{
			name:    "ERROR: IPv6 multiple double colons",
			ip:      "2001::db8::1",
			wantErr: true,
		},
		{
			name:    "ERROR: Malformed IP",
			ip:      "not.an.ip.address",
			wantErr: true,
		},
		{
			name:    "ERROR: Just numbers",
			ip:      "123456789",
			wantErr: true,
		},
		
		// Invalid CIDR Cases
		{
			name:    "ERROR: Invalid IPv4 CIDR negative",
			ip:      ipAddress,
			cidr:    subnet2,
			wantErr: true,
		},
		{
			name:    "ERROR: IPv4 CIDR too large",
			ip:      "192.168.1.1",
			cidr:    "33",
			wantErr: true,
		},
		{
			name:    "ERROR: IPv6 CIDR too large",
			ip:      ipv6Address1,
			cidr:    "129",
			wantErr: true,
		},
		{
			name:    "ERROR: CIDR non-numeric",
			ip:      "192.168.1.1",
			cidr:    "abc",
			wantErr: true,
		},
		{
			name:    "ERROR: CIDR with slash",
			ip:      "192.168.1.1",
			cidr:    "/24",
			wantErr: true,
		},
		{
			name:    "ERROR: CIDR empty string",
			ip:      "192.168.1.1",
			cidr:    "",
			wantErr: false, // Empty CIDR should be valid (no subnet)
		},
		{
			name:    "ERROR: CIDR with spaces",
			ip:      "192.168.1.1",
			cidr:    " 24 ",
			wantErr: true,
		},
		
		// Edge Cases
		{
			name:    "ERROR: IP with null byte",
			ip:      "192.168.1.1\x00",
			wantErr: true,
		},
		{
			name:    "ERROR: Very long string",
			ip:      "192.168.1.1" + string(make([]byte, 1000)),
			wantErr: true,
		},
		{
			name:    "ERROR: Unicode characters",
			ip:      "192.168.１.1", // Full-width character
			wantErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err := ValidateIPAddress(tC.ip, tC.cidr)
			if tC.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestIsIPIdentical(t *testing.T) {
	testCases := []struct {
		name    string
		ip      string
		input   string
		wantErr bool
	}{
		{
			name:  "SUCCESS: Valid IPv4",
			ip:    ipAddress3,
			input: ipAddress3,
		},
		{
			name:  "SUCCESS: Valid IPv6",
			ip:    ipv6Address1,
			input: ipv6Address1,
		},
		{
			name:    "ERROR: IPv4 not matched",
			ip:      ipAddress,
			input:   ipAddress2,
			wantErr: true,
		},
		{
			name:    "ERROR: IPv6 not matched",
			ip:      ipv6Address1,
			input:   ipv6Address2,
			wantErr: true,
		},
		{
			name:    "ERROR: IPv4 vs IPv6",
			ip:      ipAddress,
			input:   ipv6Address1,
			wantErr: true,
		},
		{
			name:    "ERROR: Invalid Config IP",
			ip:      ipAddress,
			input:   ipAddress2,
			wantErr: true,
		},
		{
			name:    "ERROR: Invalid IP",
			ip:      ipAddress,
			input:   ipAddress2,
			wantErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err := IsIPIdentical(tC.ip, tC.input)
			if tC.wantErr {
				assert.Equal(t, err, false)
			}
		})
	}
}

func TestIsInIPRange(t *testing.T) {
	testCases := []struct {
		name   string
		ip     string
		input  string
		result bool
	}{
		{
			name:   "SUCCESS: IPv4 within range",
			ip:     ipAddress4,
			input:  ipAddress3,
			result: true,
		},
		{
			name:   "ERROR: IPv4 outside range",
			ip:     ipAddress5,
			input:  ipAddress3,
			result: false,
		},
		{
			name:   "SUCCESS: IPv6 within range",
			ip:     ipv6Range1,
			input:  ipv6Address1,
			result: true,
		},
		{
			name:   "SUCCESS: IPv6 within range (same network)",
			ip:     ipv6Range1,
			input:  ipv6Address2,
			result: true,
		},
		{
			name:   "ERROR: IPv6 outside range",
			ip:     ipv6Range2,
			input:  ipv6Address1,
			result: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err := IsInIPRange(tC.ip, tC.input)
			assert.Equal(t, err, tC.result)
		})
	}
}

func TestIsIPMatch(t *testing.T) {
	testCases := []struct {
		name         string
		configIp     string
		configSubnet string
		input        string
		result       bool
	}{
		{
			name:         "SUCCESS: Default All IPv4",
			configIp:     configIp,
			configSubnet: subnet,
			input:        ipAddress3,
			result:       true,
		},
		{
			name:         "SUCCESS: Default All IPv4 no subnet",
			configIp:     configIp,
			configSubnet: "",
			input:        ipAddress3,
			result:       true,
		},
		{
			name:         "SUCCESS: Default All IPv6",
			configIp:     configIPv6_1,
			configSubnet: ipv6Subnet,
			input:        ipv6Address1,
			result:       true,
		},
		{
			name:         "SUCCESS: Default All IPv6 no subnet",
			configIp:     configIPv6_1,
			configSubnet: "",
			input:        ipv6Address1,
			result:       true,
		},
		{
			name:         "SUCCESS: IPv4 exact match",
			configIp:     configIp2,
			configSubnet: "",
			input:        ipAddress3,
			result:       true,
		},
		{
			name:         "SUCCESS: IPv6 exact match",
			configIp:     configIPv6_3,
			configSubnet: "",
			input:        ipv6Address1,
			result:       true,
		},
		{
			name:         "SUCCESS: IPv6 subnet match",
			configIp:     configIPv6_2,
			configSubnet: ipv6Subnet,
			input:        ipv6Address1,
			result:       true,
		},
		{
			name:         "SUCCESS: IPv6 subnet match with different host",
			configIp:     configIPv6_2,
			configSubnet: ipv6Subnet,
			input:        ipv6Address2,
			result:       true,
		},
		{
			name:         "ERROR: IPv4 outside range",
			configIp:     configIp3,
			configSubnet: subnet,
			input:        ipAddress3,
			result:       false,
		},
		{
			name:         "ERROR: IPv6 outside subnet",
			configIp:     configIPv6_2,
			configSubnet: ipv6Subnet,
			input:        ipv6Address3,
			result:       false,
		},
		{
			name:         "ERROR: IPv6 exact mismatch",
			configIp:     configIPv6_3,
			configSubnet: "",
			input:        ipv6Address2,
			result:       false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err := IsIPMatch(tC.configIp, tC.configSubnet, tC.input)
			assert.Equal(t, err, tC.result)
		})
	}
}
