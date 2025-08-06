package main

import (
	"net"
	"reflect"
	"testing"
)

func toCIDRs(strs []string) []*net.IPNet {
	var cidrs []*net.IPNet
	for _, s := range strs {
		n, err := parseIPOrCIDR(s)
		if err != nil {
			panic(err)
		}
		cidrs = append(cidrs, n)
	}
	return cidrs
}

func cidrsToStrings(cidrs []*net.IPNet) []string {
	var out []string
	for _, c := range cidrs {
		out = append(out, c.String())
	}
	return out
}

func TestClusterCIDRs(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Simple IPv4 merge",
			input:    []string{"192.168.0.0/25", "192.168.0.128/25"},
			expected: []string{"192.168.0.0/24"},
		},
		{
			name:     "IPv4 IPs converted and merged",
			input:    []string{"10.0.0.0", "10.0.0.1", "10.0.0.2", "10.0.0.3"},
			expected: []string{"10.0.0.0/30"},
		},
		{
			name:     "IPv6 merge",
			input:    []string{"2001:db8::/126", "2001:db8::4/126"},
			expected: []string{"2001:db8::/125"},
		},
		{
			name:     "Mixed IPv4 and IPv6",
			input:    []string{"192.0.2.0/25", "192.0.2.128/25", "2001:db8::1"},
			expected: []string{"192.0.2.0/24", "2001:db8::1/128"},
		},
		{
			name:     "Non-mergeable CIDRs",
			input:    []string{"10.0.0.0/25", "10.0.0.128/26"},
			expected: []string{"10.0.0.0/25", "10.0.0.128/26"},
		},
		{
			name:     "Single IP",
			input:    []string{"8.8.8.8"},
			expected: []string{"8.8.8.8/32"},
		},
		{
			name:     "Empty input",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ClusterCIDRs(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotStr := cidrsToStrings(got)
			wantStr := cidrsToStrings(toCIDRs(tc.expected))

			// Sort-independent comparison
			if !reflect.DeepEqual(gotStr, wantStr) {
				t.Errorf("Expected: %v\nGot:      %v", wantStr, gotStr)
			}
		})
	}
}
