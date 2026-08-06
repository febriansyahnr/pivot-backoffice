package util

import "testing"

func TestGetIPVersion(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected string
	}{
		{
			name:     "empty ip returns v4 default",
			ip:       "",
			expected: "v4",
		},
		{
			name:     "valid ipv4 localhost",
			ip:       "127.0.0.1",
			expected: "v4",
		},
		{
			name:     "valid ipv4 private",
			ip:       "192.168.1.1",
			expected: "v4",
		},
		{
			name:     "valid ipv4 public",
			ip:       "8.8.8.8",
			expected: "v4",
		},
		{
			name:     "valid ipv6 localhost",
			ip:       "::1",
			expected: "v6",
		},
		{
			name:     "valid ipv6 full",
			ip:       "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expected: "v6",
		},
		{
			name:     "valid ipv6 compressed",
			ip:       "2001:db8::1",
			expected: "v6",
		},
		{
			name:     "valid ipv6 loopback",
			ip:       "0:0:0:0:0:0:0:1",
			expected: "v6",
		},
		{
			name:     "invalid ip returns v4 default",
			ip:       "invalid",
			expected: "v4",
		},
		{
			name:     "invalid ip with partial numbers",
			ip:       "192.168.1",
			expected: "v4",
		},
		{
			name:     "invalid ip with extra octets",
			ip:       "192.168.1.1.1",
			expected: "v4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetIPVersion(tt.ip)
			if result != tt.expected {
				t.Errorf("GetIPVersion(%q) = %q, expected %q", tt.ip, result, tt.expected)
			}
		})
	}
}
