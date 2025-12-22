package netrie

import (
	"fmt"
	"net"
)

// Adder is an interface for adding IP networks or CIDR ranges to a data structure with associated names.
type Adder interface {
	AddNet(ipNet *net.IPNet, name string)
	AddCIDR(cidr string, name string) error
	Metadata() *Metadata
}

// IPLookuper defines methods to lookup and retrieve information for a given IP or IP string from a CIDR-based structure.
type IPLookuper interface {
	LookupIP(ip net.IP) string
	Lookup(ipStr string) string
	Len() int
	LenNames() int
	Metadata() *Metadata
}

// SafeIPLookuper defines methods to lookup and retrieve information for a given IP or IP string from a CIDR-based structure.
type SafeIPLookuper interface {
	SafeLookupIP(ip net.IP) (string, error)
	LookupIP(ip net.IP) string
	Lookup(ipStr string) string
	Len() int
	LenNames() int
	Metadata() *Metadata
	Close() error
}

// NewCIDRLargeIndex initializes a new CIDR trie with a root node for up to 2^32 networks.
func NewCIDRLargeIndex() *CIDRIndex[int32] {
	return newCIDRIndex[int32]()
}

// NewCIDRIndex initializes a new CIDR trie with a root node for up to 2^16 networks.
func NewCIDRIndex() *CIDRIndex[int16] {
	return newCIDRIndex[int16]()
}

// Metadata returns a reference to the Metadata object associated with the CIDRIndex.
func (idx *CIDRIndex[S]) Metadata() *Metadata {
	return &idx.meta
}

// Len returns the number of CIDRs in the trie.
func (idx *CIDRIndex[S]) Len() int {
	return idx.total
}

// LenNodes returns the number of nodes in the trie.
func (idx *CIDRIndex[S]) LenNodes() int {
	return len(idx.nodes)
}

// LenNames returns the number of different names in the trie.
func (idx *CIDRIndex[S]) LenNames() int {
	return len(idx.idByName)
}

// AddCIDR adds a CIDR with an associated id to the trie.
// Returns error if CIDR is invalid or overlaps.
func (idx *CIDRIndex[S]) AddCIDR(cidr string, name string) error {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR (%s): %v", name, cidr)
	}

	idx.AddNet(ipNet, name)

	return nil
}

// Lookup finds the id of the CIDR that contains the given IP string.
// Returns "" if no matching CIDR is found or IP is invalid.
func (idx *CIDRIndex[S]) Lookup(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "" // Invalid IP address.
	}

	return idx.LookupIP(ip)
}
