package netrie_test

import (
	"testing"

	"github.com/vearutop/netrie"
)

func TestNewCIDRTrie(t *testing.T) {
	trie := netrie.NewCIDRIndex()

	// Example usage.
	cidrs := []struct {
		cidr string
		name string
	}{
		{"192.168.1.0/24", "net1"},
		{"192.168.0.0/16", "net2"},
		{"10.0.0.0/8", "net3"},
		{"2001:db8::/32", "net4"},
	}

	for _, c := range cidrs {
		if err := trie.AddCIDR(c.cidr, c.name); err != nil {
			t.Errorf("Error adding CIDR %s: %v\n", c.cidr, err)
		}
	}

	for _, tc := range []struct {
		ip   string
		name string
	}{
		{"192.168.1.100", "net1"},
		{"192.168.2.100", "net2"},
		{"10.0.0.1", "net3"},
		{"172.16.0.1", ""},
		{"10.0.1.52", "net3"},
		{"2001:db8::1", "net4"},
		{"invalid", ""},
	} {
		name := trie.Lookup(tc.ip)
		if name != tc.name {
			t.Errorf("Lookup(%q) = %s, want %s", tc.ip, name, tc.name)
		}
	}
}
