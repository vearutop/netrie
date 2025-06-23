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
		id   int16
	}{
		{"192.168.1.0/24", 1},
		{"192.168.0.0/16", 2},
		{"10.0.0.0/8", 3},
	}

	for _, c := range cidrs {
		if err := trie.AddCIDR(c.cidr, c.id); err != nil {
			t.Errorf("Error adding CIDR %s: %v\n", c.cidr, err)
		}
	}

	for _, tc := range []struct {
		ip string
		id int16
	}{
		{"192.168.1.100", 1},
		{"192.168.2.100", 2},
		{"10.0.0.1", 3},
		{"172.16.0.1", -1},
		{"10.0.1.52", 3},
		{"invalid", -1},
	} {
		id := trie.Lookup(tc.ip)
		if id != tc.id {
			t.Errorf("Lookup(%q) = %d, want %d", tc.ip, id, tc.id)
		}
	}
}
