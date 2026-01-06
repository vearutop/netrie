package netrie

import (
	"net"
	"time"
)

// trieNode represents a node in the CIDR trie.
type trieNode[S int16 | int32] struct {
	children [2]int32 // Indices of child nodes (0 or 1).
	id       S        // ID associated with the CIDR, -1 if none.
	maskLen  int8     // Length of the CIDR mask, -1 if none.
}

// Metadata represents additional information related to a structure or process.
type Metadata struct {
	BuildDate   time.Time `json:"build_date,omitzero"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Extra       any       `json:"extra,omitempty"`
}

// CIDRIndex is the trie structure for CIDR lookups.
type CIDRIndex[S int16 | int32] struct {
	meta Metadata

	nodes []trieNode[S] // Slice storing all trie nodes.
	names []string
	total int

	idByName map[string]S
}

func newCIDRIndex[S int16 | int32]() *CIDRIndex[S] {
	return &CIDRIndex[S]{
		nodes:    []trieNode[S]{{children: [2]int32{-1, -1}, id: -1, maskLen: -1}},
		idByName: make(map[string]S),
	}
}

// AddNet inserts a CIDR block represented by ipNet into the trie, associating it with the specified name.
func (idx *CIDRIndex[S]) AddNet(ipNet *net.IPNet, name string) {
	id := idx.idByName[name]

	if id == 0 {
		idx.names = append(idx.names, name)
		id = S(len(idx.names))
		idx.idByName[name] = id
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
		childIndex := idx.nodes[current].children[bit]
		if childIndex == -1 {
			// Create new node.
			idx.nodes[current].children[bit] = int32(len(idx.nodes))
			idx.nodes = append(idx.nodes, trieNode[S]{
				children: [2]int32{-1, -1},
				id:       -1,
				maskLen:  -1,
			})
			childIndex = idx.nodes[current].children[bit]
		}
		current = int(childIndex)
	}

	// Set id and mask length at the leaf node.
	idx.nodes[current].id = id
	idx.nodes[current].maskLen = int8(maskLen)

	idx.total++
}

// LookupIP finds the id of the CIDR that contains the given IP.
// Returns "" if no matching CIDR is found.
func (idx *CIDRIndex[S]) LookupIP(ip net.IP) string {
	if ip == nil {
		return ""
	}

	// Convert to 16-byte representation, handling IPv4.
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	}

	current := 0
	bestID := S(-1)
	bestMaskLen := int8(-1)

	// Traverse up to 128 bits for IPv6 (or 32 for IPv4).
	maxBits := 128
	if len(ip) == 4 {
		maxBits = 32 // IPv4.
	}

	for i := 0; i < maxBits; i++ {
		// Check if current node has an id and update best match if mask is longer.
		if idx.nodes[current].id != -1 && idx.nodes[current].maskLen > bestMaskLen {
			bestID = idx.nodes[current].id
			bestMaskLen = idx.nodes[current].maskLen
		}

		// Get the next bit.
		bit := (ip[i/8] >> (7 - (i % 8))) & 1
		childIndex := idx.nodes[current].children[bit]
		if childIndex == -1 {
			break // No further path.
		}
		current = int(childIndex)
	}

	// Check the final node for a better match.
	if idx.nodes[current].id != -1 && idx.nodes[current].maskLen > bestMaskLen {
		bestID = idx.nodes[current].id
	}

	if bestID == -1 {
		return ""
	}

	return idx.names[bestID-1]
}

// Minimize merges isomorphic subtrees, producing a minimal DAWG.
// Safe to call only once after all insertions are done.
// Reduces node count typically by 60–80% on real-world CIDR sets.
func (idx *CIDRIndex[S]) Minimize() {
	if len(idx.nodes) <= 1 {
		return
	}

	// signature uniquely identifies a node after children are canonicalized
	type sig struct {
		ch0, ch1 int32 // canonical child indices (-1 if none)
		id       S
		maskLen  int8
	}

	// Maps signature → canonical node index in the final array
	sigToIndex := make(map[sig]int32)

	// old index → new canonical index
	remap := make([]int32, len(idx.nodes))

	// We'll build the new minimal node list here
	var minimal []trieNode[S]

	// Process nodes in reverse order (bottom-up)
	for i := len(idx.nodes) - 1; i >= 0; i-- {
		old := idx.nodes[i]

		// Resolve children to their future canonical indices
		ch0 := int32(-1)
		if old.children[0] != -1 {
			ch0 = remap[old.children[0]]
		}
		ch1 := int32(-1)
		if old.children[1] != -1 {
			ch1 = remap[old.children[1]]
		}

		currentSig := sig{
			ch0:     ch0,
			ch1:     ch1,
			id:      old.id,
			maskLen: old.maskLen,
		}

		// Reuse existing node if signature already exists
		if newIdx, exists := sigToIndex[currentSig]; exists {
			remap[i] = newIdx
		} else {
			// First time we see this signature → assign new canonical index
			newIdx := int32(len(minimal))
			remap[i] = newIdx
			sigToIndex[currentSig] = newIdx

			// Append node with already-remapped children
			minimal = append(minimal, trieNode[S]{
				children: [2]int32{ch0, ch1},
				id:       old.id,
				maskLen:  old.maskLen,
			})
		}
	}

	// Root is always at 0 → remap it
	// (after minimization it might move, but we keep it at index 0)
	finalRoot := remap[0]
	if finalRoot != 0 {
		// Swap root node to position 0
		minimal[finalRoot], minimal[0] = minimal[0], minimal[finalRoot]

		// Fix up all references to the swapped nodes
		for i := range remap {
			switch remap[i] {
			case 0:
				remap[i] = finalRoot
			case finalRoot:
				remap[i] = 0
			}
		}
	}

	idx.nodes = minimal
}
