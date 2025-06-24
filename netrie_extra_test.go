package netrie

import (
	"fmt"
	"net"
	"testing"
)

// TestNewCIDRIndex verifies that a new CIDRIndex is initialized correctly.
func TestNewCIDRIndex(t *testing.T) {
	idx := NewCIDRIndex()
	if idx == nil {
		t.Fatal("NewCIDRIndex returned nil")
	}
	if len(idx.nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(idx.nodes))
	}
	if idx.nodes[0].id != -1 || idx.nodes[0].maskLen != -1 {
		t.Errorf("Root node not initialized correctly: id=%d, maskLen=%d", idx.nodes[0].id, idx.nodes[0].maskLen)
	}
	if idx.nodes[0].children != [2]int32{-1, -1} {
		t.Errorf("Root node children not initialized: %v", idx.nodes[0].children)
	}
	if idx.Len() != 0 {
		t.Errorf("Expected Len() to be 0, got %d", idx.Len())
	}
}

// TestAddCIDR_IPv4 tests adding valid IPv4 CIDRs and checks for correct behavior.
func TestAddCIDR_IPv4(t *testing.T) {
	idx := NewCIDRIndex()
	tests := []struct {
		cidr string
		name string
		err  error
	}{
		{"192.168.1.0/24", "net1", nil},
		{"10.0.0.0/8", "net2", nil},
		{"172.16.0.0/16", "net3", nil},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.cidr, tt.name), func(t *testing.T) {
			err := idx.AddCIDR(tt.cidr, tt.name)
			if err != tt.err {
				t.Errorf("Expected error %v, got %v", tt.err, err)
			}
			if idx.Len() != i+1 {
				t.Errorf("Expected Len() to be %d, got %d", i+1, idx.Len())
			}
		})
	}
}

// TestAddCIDR_IPv6 tests adding valid IPv6 CIDRs.
func TestAddCIDR_IPv6(t *testing.T) {
	idx := NewCIDRIndex()
	tests := []struct {
		cidr string
		name string
		err  error
	}{
		{"2001:db8::/32", "net1", nil},
		{"2001:db8:1::/48", "net2", nil},
		{"2001:db8:2::/64", "net3", nil},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.cidr, tt.name), func(t *testing.T) {
			err := idx.AddCIDR(tt.cidr, tt.name)
			if err != tt.err {
				t.Errorf("Expected error %v, got %v", tt.err, err)
			}
			if idx.Len() != i+1 {
				t.Errorf("Expected Len() to be %d, got %d", i+1, idx.Len())
			}
		})
	}
}

// TestAddCIDR_Overlap tests adding overlapping CIDRs and expects errors.
func TestAddCIDR_Overlap(t *testing.T) {
	idx := NewCIDRIndex()
	tests := []struct {
		cidr string
		name string
		err  error
	}{
		{"192.168.1.0/24", "net1", nil},
		{"192.168.1.0/24", "net2", fmt.Errorf("%w net2 with net1: 192.168.1.0/24", ErrOverlap)},
		{"192.168.1.128/25", "net3", fmt.Errorf("%w net3 with net1: 192.168.1.128/25", ErrOverlap)},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.cidr, tt.name), func(t *testing.T) {
			err := idx.AddCIDR(tt.cidr, tt.name)
			if err != nil && tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("Expected error %v, got %v", tt.err, err)
				}
			} else if err != tt.err {
				t.Errorf("Expected error %v, got %v", tt.err, err)
			}
		})
	}
}

// TestAddCIDR_Invalid tests adding invalid CIDRs.
func TestAddCIDR_Invalid(t *testing.T) {
	idx := NewCIDRIndex()
	tests := []struct {
		cidr string
		name string
		err  bool
	}{
		{"256.256.256.256/32", "invalid1", true},
		{"192.168.1.0/33", "invalid2", true},
		{"2001:db8::/129", "invalid3", true},
		{"not-a-cidr", "invalid4", true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.cidr, tt.name), func(t *testing.T) {
			err := idx.AddCIDR(tt.cidr, tt.name)
			if (err != nil) != tt.err {
				t.Errorf("Expected error %v, got %v", tt.err, err)
			}
			if err == nil && idx.Len() != 0 {
				t.Errorf("Expected Len() to be 0, got %d", idx.Len())
			}
		})
	}
}

// TestLookup_IPv4 tests IP lookups for IPv4 addresses.
func TestLookup_IPv4(t *testing.T) {
	idx := NewCIDRIndex()
	// Add some CIDRs.
	cidrs := []struct{ cidr, name string }{
		{"192.168.1.0/24", "net1"},
		{"10.0.0.0/8", "net2"},
		{"172.16.0.0/12", "net3"},
	}
	for _, c := range cidrs {
		if err := idx.AddCIDR(c.cidr, c.name); err != nil {
			t.Fatalf("Failed to add CIDR %s: %v", c.cidr, err)
		}
	}

	tests := []struct {
		ip       string
		expected string
	}{
		{"192.168.1.100", "net1"},
		{"10.20.30.40", "net2"},
		{"172.20.1.1", "net3"},
		{"8.8.8.8", ""},    // No match.
		{"invalid-ip", ""}, // Invalid IP.
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := idx.Lookup(tt.ip)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestLookup_IPv6 tests IP lookups for IPv6 addresses.
func TestLookup_IPv6(t *testing.T) {
	idx := NewCIDRIndex()
	// Add some CIDRs.
	cidrs := []struct{ cidr, name string }{
		{"2001:db8::/32", "net1"},
		{"2001:db8:1::/48", "net2"},
		{"2001:db8:2::/64", "net3"},
	}
	for _, c := range cidrs {
		if err := idx.AddCIDR(c.cidr, c.name); err != nil {
			t.Fatalf("Failed to add CIDR %s: %v", c.cidr, err)
		}
	}

	tests := []struct {
		ip       string
		expected string
	}{
		{"2001:db8::1", "net1"},
		{"2001:db8:1::abcd", "net2"},
		{"2001:db8:2::1234", "net3"},
		{"2001:db9::1", ""}, // No match.
		{"invalid-ip", ""},  // Invalid IP.
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := idx.Lookup(tt.ip)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestLookupIP tests LookupIP with net.IP directly.
func TestLookupIP(t *testing.T) {
	idx := NewCIDRIndex()
	// Add CIDRs.
	if err := idx.AddCIDR("192.168.1.0/24", "net1"); err != nil {
		t.Fatal(err)
	}
	if err := idx.AddCIDR("2001:db8::/32", "net2"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		ip       net.IP
		expected string
	}{
		{net.ParseIP("192.168.1.100"), "net1"},
		{net.ParseIP("2001:db8::1"), "net2"},
		{net.ParseIP("8.8.8.8"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.ip.String(), func(t *testing.T) {
			result := idx.LookupIP(tt.ip)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestLongestPrefixMatch tests that the longest prefix match is returned.
func TestLongestPrefixMatch(t *testing.T) {
	idx := NewCIDRIndex()
	// Add overlapping CIDRs with different mask lengths.
	cidrs := []struct{ cidr, name string }{
		{"192.168.0.0/16", "net1"},
		{"192.168.1.0/24", "net2"},
		{"192.168.1.128/25", "net3"},
	}
	for _, c := range cidrs {
		if err := idx.AddCIDR(c.cidr, c.name); err != nil {
			t.Fatalf("Failed to add CIDR %s: %v", c.cidr, err)
		}
	}

	tests := []struct {
		ip       string
		expected string
	}{
		{"192.168.1.129", "net3"}, // Matches /25.
		{"192.168.1.1", "net2"},   // Matches /24.
		{"192.168.2.1", "net1"},   // Matches /16.
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := idx.Lookup(tt.ip)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestEmptyTrieLookup tests lookups in an empty trie.
func TestEmptyTrieLookup(t *testing.T) {
	idx := NewCIDRIndex()
	tests := []string{
		"192.168.1.1",
		"2001:db8::1",
		"invalid-ip",
	}
	for _, ip := range tests {
		t.Run(ip, func(t *testing.T) {
			result := idx.Lookup(ip)
			if result != "" {
				t.Errorf("Expected empty string, got %q", result)
			}
		})
	}
}

// TestNameReuse tests that adding CIDRs with the same name reuses the name ID.
func TestNameReuse(t *testing.T) {
	idx := NewCIDRIndex()
	if err := idx.AddCIDR("192.168.1.0/24", "net1"); err != nil {
		t.Fatal(err)
	}
	if err := idx.AddCIDR("192.168.2.0/24", "net1"); err != nil {
		t.Fatal(err)
	}
	if len(idx.names) != 1 {
		t.Errorf("Expected 1 name, got %d", len(idx.names))
	}
	if idx.names[0] != "net1" {
		t.Errorf("Expected name 'net1', got %q", idx.names[0])
	}
	if idx.Len() != 2 {
		t.Errorf("Expected Len() to be 2, got %d", idx.Len())
	}
}
