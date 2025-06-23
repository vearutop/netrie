package netrie

import (
	"fmt"
	"net"
)

type errString string

func (e errString) Error() string {
	return string(e)
}

const (
	ErrOverlap = errString("overlap")
)

// trieNode represents a node in the CIDR trie.
type trieNode struct {
	children [2]int32 // Indices of child nodes (0 or 1).
	id       int16    // ID associated with the CIDR, -1 if none.
	maskLen  int8     // Length of the CIDR mask, -1 if none.
}

// CIDRIndex is the trie structure for CIDR lookups.
type CIDRIndex struct {
	nodes []trieNode // Slice storing all trie nodes.
	total int
}

// NewCIDRIndex initializes a new CIDR trie with a root node.
func NewCIDRIndex() *CIDRIndex {
	return &CIDRIndex{
		nodes: []trieNode{{children: [2]int32{-1, -1}, id: -1, maskLen: -1}},
	}
}

// Len returns the number of CIDRs in the trie.
func (t *CIDRIndex) Len() int {
	return t.total
}

// AddCIDR adds a CIDR with an associated id to the trie.
// Returns error if CIDR is invalid or overlaps.
func (t *CIDRIndex) AddCIDR(cidr string, id int16) error {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %v", cidr)
	}

	// Get 16-byte IP representation (IPv4 or IPv6).
	ip := ipNet.IP
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4 // Convert IPv4 to 4-byte representation.
	}

	maskLen, _ := ipNet.Mask.Size()
	current := 0 // Start at root node.

	// Traverse or build the trie for each bit in the mask.
	for i := 0; i < maskLen; i++ {
		bit := (ip[i/8] >> (7 - (i % 8))) & 1
		childIndex := t.nodes[current].children[bit]
		if childIndex == -1 {
			// Create new node.
			t.nodes[current].children[bit] = int32(len(t.nodes))
			t.nodes = append(t.nodes, trieNode{
				children: [2]int32{-1, -1},
				id:       -1,
				maskLen:  -1,
			})
			childIndex = t.nodes[current].children[bit]
		}
		current = int(childIndex)
	}

	// Set id and mask length at the leaf node.
	if t.nodes[current].id != -1 {
		return fmt.Errorf("%w: %s", ErrOverlap, cidr)
	}
	t.nodes[current].id = id
	t.nodes[current].maskLen = int8(maskLen)

	t.total++

	return nil
}

// Lookup finds the id of the CIDR that contains the given IP string.
// Returns -1 if no matching CIDR is found or IP is invalid.
func (t *CIDRIndex) Lookup(ipStr string) int16 {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return -1 // Invalid IP address.
	}

	return t.LookupIP(ip)
}

// LookupIP finds the id of the CIDR that contains the given IP.
// Returns -1 if no matching CIDR is found.
func (t *CIDRIndex) LookupIP(ip net.IP) int16 {
	// Convert to 16-byte representation, handling IPv4.
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	}

	current := 0
	bestID := int16(-1)
	bestMaskLen := int8(-1)

	// Traverse up to 128 bits for IPv6 (or 32 for IPv4).
	maxBits := 128
	if len(ip) == 4 {
		maxBits = 32 // IPv4.
	}

	for i := 0; i < maxBits; i++ {
		// Check if current node has an id and update best match if mask is longer.
		if t.nodes[current].id != -1 && t.nodes[current].maskLen > bestMaskLen {
			bestID = t.nodes[current].id
			bestMaskLen = t.nodes[current].maskLen
		}

		// Get the next bit.
		bit := (ip[i/8] >> (7 - (i % 8))) & 1
		childIndex := t.nodes[current].children[bit]
		if childIndex == -1 {
			break // No further path.
		}
		current = int(childIndex)
	}

	// Check the final node for a better match.
	if t.nodes[current].id != -1 && t.nodes[current].maskLen > bestMaskLen {
		bestID = t.nodes[current].id
	}

	return bestID
}
