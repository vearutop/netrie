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

// trieNode represents a node in the trie, stored in a flat slice.
type trieNode struct {
	children [2]int32 // Indices of left (0) and right (1) children in the nodes slice; -1 means no child.
	id       int16    // id of the CIDR if this node is a leaf; -1 means no id.
	maskLen  int8     // Mask length of the CIDR at this node; -1 if not a leaf.
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

func (t *CIDRIndex) Len() int {
	return t.total
}

// AddCIDR adds a CIDR with an associated id to the trie.
// Returns error if CIDR is invalid.
func (t *CIDRIndex) AddCIDR(cidr string, id int16) error {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %v", cidr)
	}

	ip := ipNet.IP.To4()
	if ip == nil {
		return fmt.Errorf("only IPv4 supported: %s", cidr)
	}

	maskLen, _ := ipNet.Mask.Size()
	current := 0 // Start at root node.

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

func (t *CIDRIndex) Lookup(ipStr string) int16 {
	ip := net.ParseIP(ipStr).To4()
	if ip == nil {
		return -1 // Invalid or non-IPv4 address.
	}

	return t.LookupIP(ip)
}

// LookupIP finds the id of the CIDR that contains the given IP.
// Returns -1 if no matching CIDR is found.
func (t *CIDRIndex) LookupIP(ip net.IP) int16 {
	current := 0
	bestID := int16(-1)
	bestMaskLen := int8(-1)

	for i := 0; i < 32; i++ {
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
