package server

import (
	"testing"
)

func TestIsInWhiteList(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		whiteList []string
		expected  bool
	}{
		{
			name:      "Empty whitelist",
			ip:        "192.168.1.1",
			whiteList: []string{},
			expected:  false,
		},
		{
			name:      "Exact match",
			ip:        "192.168.1.1",
			whiteList: []string{"10.0.0.1", "192.168.1.1"},
			expected:  true,
		},
		{
			name:      "Single asterisk prefix match",
			ip:        "192.168.1.42",
			whiteList: []string{"192.168.*"},
			expected:  true,
		},
		{
			name:      "Multi-level asterisk match",
			ip:        "10.20.30.40",
			whiteList: []string{"10.*"},
			expected:  true,
		},
		{
			name:      "No match in list",
			ip:        "172.16.0.1",
			whiteList: []string{"192.168.*", "10.0.0.1"},
			expected:  false,
		},
		{
			name:      "Asterisk not at end",
			ip:        "192.168.1.100",
			whiteList: []string{"192.*.1.*"},
			expected:  false,
		},
		{
			name:      "Prefix longer than IP",
			ip:        "192.168.1",
			whiteList: []string{"192.168.1.*"},
			expected:  false,
		},
		{
			name:      "Partial match but continues",
			ip:        "192.168.2.1",
			whiteList: []string{"192.168.*", "192.168.2.*"},
			expected:  true,
		},
		{
			name:      "Match after earlier mismatch",
			ip:        "10.1.2.3",
			whiteList: []string{"192.168.*", "10.*"},
			expected:  true,
		},
		{
			name:      "Invalid asterisk position skipped",
			ip:        "192.168.1.1",
			whiteList: []string{"192.168.*.1"},
			expected:  false,
		},
		{
			name:      "Short IP with valid pattern",
			ip:        "1.2.3.4",
			whiteList: []string{"1.2.*"},
			expected:  true,
		},
		{
			name:      "IPv6 should not match",
			ip:        "2001:db8::1",
			whiteList: []string{"192.168.*"},
			expected:  false,
		},
		{
			name:      "Invalid IP format",
			ip:        "invalid-ip",
			whiteList: []string{"192.168.*"},
			expected:  false,
		},
		{
			name:      "Multiple wildcards in list",
			ip:        "10.20.30.40",
			whiteList: []string{"192.168.*", "10.*", "172.16.*"},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInWhiteList(tt.ip, tt.whiteList)
			if result != tt.expected {
				t.Errorf(
					"expected %v, got %v for IP: %s, whitelist: %v",
					tt.expected,
					result,
					tt.ip,
					tt.whiteList,
				)
			}
		})
	}
}
