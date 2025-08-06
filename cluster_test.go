package netrie

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
		err      bool
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
		{
			name: "IPv4 sequential IPs",
			input: []string{
				"192.168.1.0",
				"192.168.1.1",
				"192.168.1.2",
				"192.168.1.3",
			},
			expected: []string{"192.168.1.0/30"},
			err:      false,
		},
		{
			name: "IPv4 non-sequential IPs",
			input: []string{
				"192.168.1.0",
				"192.168.1.2",
				"192.168.1.4",
			},
			expected: []string{
				"192.168.1.0/32",
				"192.168.1.2/32",
				"192.168.1.4/32",
			},
			err: false,
		},
		{
			name: "IPv4 mixed IPs and CIDRs",
			input: []string{
				"192.168.1.0/30",
				"192.168.1.4",
				"192.168.1.5",
			},
			expected: []string{
				"192.168.1.0/30",
				"192.168.1.4/31",
			},
			err: false,
		},
		{
			name: "IPv6 sequential IPs",
			input: []string{
				"2001:db8::1",
				"2001:db8::2",
				"2001:db8::3",
			},
			expected: []string{"2001:db8::/126"},
			err:      false,
		},
		{
			name: "IPv6 non-sequential IPs",
			input: []string{
				"2001:db8::1",
				"2001:db8::3",
				"2001:db8::5",
			},
			expected: []string{
				"2001:db8::1/128",
				"2001:db8::3/128",
				"2001:db8::5/128",
			},
			err: false,
		},
		{
			name: "Mixed IPv4 and IPv6",
			input: []string{
				"192.168.1.0",
				"192.168.1.1",
				"2001:db8::1",
				"2001:db8::2",
			},
			expected: []string{
				"192.168.1.0/31",
				"2001:db8::/127",
			},
			err: false,
		},
		{
			name:     "Empty input",
			input:    []string{},
			expected: []string{},
			err:      false,
		},
		{
			name: "Invalid IP",
			input: []string{
				"192.168.1.0",
				"invalid.ip",
			},
			expected: nil,
			err:      true,
		},
		{
			name: "Large IPv4 CIDR",
			input: []string{
				"192.168.0.0/24",
			},
			expected: []string{"192.168.0.0/24"},
			err:      false,
		},
		{
			name: "Overlapping IPv4 CIDRs",
			input: []string{
				"192.168.1.0/30",
				"192.168.1.0/31",
			},
			expected: []string{"192.168.1.0/30"},
			err:      false,
		},
		{
			name: "IPv4 full range",
			input: []string{
				"0.0.0.0/1",
				"128.0.0.0/1",
			},
			expected: []string{"0.0.0.0/0"},
			err:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ClusterCIDRs(tc.input)
			if tc.err {
				if err == nil {
					t.Fatalf("ClusterCIDRs(%v) = %v, wanted error", tc.input, got)
				}
			}

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
