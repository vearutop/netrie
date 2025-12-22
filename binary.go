// Package netrie implements high performance CIDR index.
package netrie

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// MarshalBinary for trieNode (from previous implementation).
func (n *trieNode[S]) MarshalBinary() ([]byte, error) {
	var (
		s    S
		data []byte
	)

	switch any(s).(type) {
	case int16:
		data = make([]byte, 11)
		binary.BigEndian.PutUint32(data[0:4], uint32(n.children[0]))
		binary.BigEndian.PutUint32(data[4:8], uint32(n.children[1]))

		binary.BigEndian.PutUint16(data[8:10], uint16(n.id))
		data[10] = byte(n.maskLen)
	case int32:
		data = make([]byte, 13)
		binary.BigEndian.PutUint32(data[0:4], uint32(n.children[0]))
		binary.BigEndian.PutUint32(data[4:8], uint32(n.children[1]))

		binary.BigEndian.PutUint32(data[8:12], uint32(n.id))
		data[12] = byte(n.maskLen)
	}

	return data, nil
}

// UnmarshalBinary for trieNode (from previous implementation).
func (n *trieNode[S]) UnmarshalBinary(data []byte) error {
	var s S

	switch any(s).(type) {
	case int16:
		if len(data) != 11 {
			return fmt.Errorf("insufficient data: got %d bytes, need 11", len(data))
		}
		n.children[0] = int32(binary.BigEndian.Uint32(data[0:4]))
		n.children[1] = int32(binary.BigEndian.Uint32(data[4:8]))

		n.id = S(binary.BigEndian.Uint16(data[8:10]))
		n.maskLen = int8(data[10])
	case int32:
		if len(data) != 13 {
			return fmt.Errorf("insufficient data: got %d bytes, need 11", len(data))
		}
		n.children[0] = int32(binary.BigEndian.Uint32(data[0:4]))
		n.children[1] = int32(binary.BigEndian.Uint32(data[4:8]))

		n.id = S(binary.BigEndian.Uint32(data[8:12]))
		n.maskLen = int8(data[12])
	}

	return nil
}

// Save writes the CIDRIndex data to the given io.Writer, including metadata, nodes, and associated names.
func (idx *CIDRIndex[S]) Save(w io.Writer) error {
	var s S

	// Write header: version (int32), total (int32), nodesLen (int32), namesLen (int32)
	header := make([]byte, 16)

	ver := uint32(1) // Version 1.

	// Switch bit 31 to indicate large namespace.
	if _, ok := any(s).(int32); ok {
		ver |= 1 << 31
	}

	binary.BigEndian.PutUint32(header[0:4], ver)
	binary.BigEndian.PutUint32(header[4:8], uint32(idx.total))
	binary.BigEndian.PutUint32(header[8:12], uint32(len(idx.nodes)))
	binary.BigEndian.PutUint32(header[12:16], uint32(len(idx.names)))
	if _, err := w.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write the .Metadata field as JSON with its length as a prefix
	metadataJSON, err := json.Marshal(idx.meta)
	if err != nil {
		return fmt.Errorf("failed to encode .Metadata: %w", err)
	}
	metadataLenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(metadataLenBuf, uint32(len(metadataJSON)))

	if _, err := w.Write(metadataLenBuf); err != nil {
		return fmt.Errorf("failed to write .Metadata length: %w", err)
	}

	if _, err := w.Write(metadataJSON); err != nil {
		return fmt.Errorf("failed to write .Metadata: %w", err)
	}

	switch any(s).(type) {
	case int16:
		// Write nodes
		nodeBuf := make([]byte, 11) // Reusable buffer for each node
		for i, node := range idx.nodes {
			nodeData, err := node.MarshalBinary()
			if err != nil {
				return fmt.Errorf("failed to marshal node %d: %w", i, err)
			}
			copy(nodeBuf, nodeData)
			if _, err := w.Write(nodeBuf); err != nil {
				return fmt.Errorf("failed to write node %d: %w", i, err)
			}
		}
	case int32:
		// Write nodes
		nodeBuf := make([]byte, 13) // Reusable buffer for each node
		for i, node := range idx.nodes {
			nodeData, err := node.MarshalBinary()
			if err != nil {
				return fmt.Errorf("failed to marshal node %d: %w", i, err)
			}
			copy(nodeBuf, nodeData)
			if _, err := w.Write(nodeBuf); err != nil {
				return fmt.Errorf("failed to write node %d: %w", i, err)
			}
		}
	}

	// Write names
	for i, name := range idx.names {
		// Write string length (int32)
		nameLenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(nameLenBuf, uint32(len(name)))
		if _, err := w.Write(nameLenBuf); err != nil {
			return fmt.Errorf("failed to write name %d length: %w", i, err)
		}
		// Write string bytes
		if _, err := w.Write([]byte(name)); err != nil {
			return fmt.Errorf("failed to write name %d: %w", i, err)
		}
	}

	return nil
}

// SaveToFile saves the CIDRIndex to a file.
func (idx *CIDRIndex[S]) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create file to save index: %w", err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)

	if err := idx.Save(w); err != nil {
		return fmt.Errorf("save file: %w", err)
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush file: %w", err)
	}

	return nil
}

type hd struct {
	ver         uint32
	total       uint32
	nodesLen    uint32
	namesLen    uint32
	metadataLen uint32
	meta        Metadata

	hasLargeNamespace bool
	nodeSize          int64
}

func (h *hd) UnmarshalBinary(data []byte) error {
	if len(data) != 20 {
		return fmt.Errorf("insufficient data: got %d bytes, need 20", len(data))
	}

	h.ver = binary.BigEndian.Uint32(data[0:4])
	h.total = binary.BigEndian.Uint32(data[4:8])
	h.nodesLen = binary.BigEndian.Uint32(data[8:12])
	h.namesLen = binary.BigEndian.Uint32(data[12:16])
	h.metadataLen = binary.BigEndian.Uint32(data[16:20])

	h.hasLargeNamespace = (h.ver & (1 << 31)) != 0
	h.ver &^= 1 << 31 // Remove large namespace flag.

	if h.hasLargeNamespace {
		h.nodeSize = 13
	} else {
		h.nodeSize = 11
	}

	if h.ver != 1 {
		return fmt.Errorf("invalid version: %d", h.ver)
	}

	return nil
}

// Load initializes and returns an IPLookuper by reading and parsing data from the provided io.Reader.
// Returns an error if the input data is invalid or the operation fails.
func Load(r io.Reader) (IPLookuper, error) {
	// Read header: version (uint32), total (uint32), nodesLen (uint32), namesLen (uint32), metadataLen (uint32)
	header := make([]byte, 20)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	h := hd{}
	if err := h.UnmarshalBinary(header); err != nil {
		return nil, fmt.Errorf("unmarshal header: %w", err)
	}

	if h.metadataLen > 0 {
		metadataBuf := make([]byte, h.metadataLen)

		if _, err := io.ReadFull(r, metadataBuf); err != nil {
			return nil, fmt.Errorf("read metadata: %w", err)
		}

		if err := json.Unmarshal(metadataBuf, &h.meta); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	if h.hasLargeNamespace {
		idx := NewCIDRLargeIndex()

		if err := idx.load(h, r); err != nil {
			return nil, err
		}

		return idx, nil
	}

	idx := NewCIDRIndex()

	if err := idx.load(h, r); err != nil {
		return nil, err
	}

	return idx, nil
}

func (idx *CIDRIndex[S]) load(h hd, r io.Reader) error {
	idx.meta = h.meta

	// Initialize CIDRIndex fields
	idx.total = int(h.total)
	idx.nodes = make([]trieNode[S], h.nodesLen)
	idx.names = make([]string, h.namesLen)

	// Read nodes
	nodeBuf := make([]byte, h.nodeSize)
	for i := 0; i < int(h.nodesLen); i++ {
		if _, err := io.ReadFull(r, nodeBuf); err != nil {
			return fmt.Errorf("read node %d: %w", i, err)
		}
		if err := idx.nodes[i].UnmarshalBinary(nodeBuf); err != nil {
			return fmt.Errorf("unmarshal node %d: %w", i, err)
		}
	}

	// Read names
	for i := 0; i < int(h.namesLen); i++ {
		// Read string length (int32)
		nameLenBuf := make([]byte, 4)
		if _, err := io.ReadFull(r, nameLenBuf); err != nil {
			return fmt.Errorf("read name %d length: %w", i, err)
		}
		nameLen := int(binary.BigEndian.Uint32(nameLenBuf))

		// Read string bytes
		nameBuf := make([]byte, nameLen)
		if _, err := io.ReadFull(r, nameBuf); err != nil {
			return fmt.Errorf("read name %d len %d: %w", i, nameLen, err)
		}
		name := string(nameBuf)
		idx.names[i] = name
		idx.idByName[name] = S(i)
	}

	return nil
}

// LoadFromFile loads the entire CIDRIndex from a file to memory.
func LoadFromFile(filename string) (IPLookuper, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	r := bufio.NewReader(file)

	return Load(r)
}
