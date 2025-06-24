package netrie

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// MarshalBinary for trieNode (from previous implementation).
func (n *trieNode) MarshalBinary() ([]byte, error) {
	data := make([]byte, 11)
	binary.BigEndian.PutUint32(data[0:4], uint32(n.children[0]))
	binary.BigEndian.PutUint32(data[4:8], uint32(n.children[1]))
	binary.BigEndian.PutUint16(data[8:10], uint16(n.id))
	data[10] = byte(n.maskLen)
	return data, nil
}

// UnmarshalBinary for trieNode (from previous implementation).
func (n *trieNode) UnmarshalBinary(data []byte) error {
	if len(data) < 11 {
		return fmt.Errorf("insufficient data: got %d bytes, need 11", len(data))
	}
	n.children[0] = int32(binary.BigEndian.Uint32(data[0:4]))
	n.children[1] = int32(binary.BigEndian.Uint32(data[4:8]))
	n.id = int16(binary.BigEndian.Uint16(data[8:10]))
	n.maskLen = int8(data[10])
	return nil
}

// SaveToFile saves the CIDRIndex to a file.
func (idx *CIDRIndex) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write header: total (int32), nodesLen (int32), namesLen (int32)
	header := make([]byte, 12)
	binary.BigEndian.PutUint32(header[0:4], uint32(idx.total))
	binary.BigEndian.PutUint32(header[4:8], uint32(len(idx.nodes)))
	binary.BigEndian.PutUint32(header[8:12], uint32(len(idx.names)))
	if _, err := file.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write nodes
	nodeBuf := make([]byte, 11) // Reusable buffer for each node
	for i, node := range idx.nodes {
		nodeData, err := node.MarshalBinary()
		if err != nil {
			return fmt.Errorf("failed to marshal node %d: %w", i, err)
		}
		copy(nodeBuf, nodeData)
		if _, err := file.Write(nodeBuf); err != nil {
			return fmt.Errorf("failed to write node %d: %w", i, err)
		}
	}

	// Write names
	for i, name := range idx.names {
		// Write string length (int32)
		nameLenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(nameLenBuf, uint32(len(name)))
		if _, err := file.Write(nameLenBuf); err != nil {
			return fmt.Errorf("failed to write name %d length: %w", i, err)
		}
		// Write string bytes
		if _, err := file.Write([]byte(name)); err != nil {
			return fmt.Errorf("failed to write name %d: %w", i, err)
		}
	}

	return nil
}

// LoadFromFile loads the CIDRIndex from a file.
func (idx *CIDRIndex) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read header: total (int32), nodesLen (int32), namesLen (int32)
	header := make([]byte, 12)
	if _, err := io.ReadFull(file, header); err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}
	total := int(binary.BigEndian.Uint32(header[0:4]))
	nodesLen := int(binary.BigEndian.Uint32(header[4:8]))
	namesLen := int(binary.BigEndian.Uint32(header[8:12]))

	// Initialize CIDRIndex fields
	idx.total = total
	idx.nodes = make([]trieNode, nodesLen)
	idx.names = make([]string, namesLen)

	// Read nodes
	nodeBuf := make([]byte, 11)
	for i := 0; i < nodesLen; i++ {
		if _, err := io.ReadFull(file, nodeBuf); err != nil {
			return fmt.Errorf("failed to read node %d: %w", i, err)
		}
		if err := idx.nodes[i].UnmarshalBinary(nodeBuf); err != nil {
			return fmt.Errorf("failed to unmarshal node %d: %w", i, err)
		}
	}

	// Read names
	for i := 0; i < namesLen; i++ {
		// Read string length (int32)
		nameLenBuf := make([]byte, 4)
		if _, err := io.ReadFull(file, nameLenBuf); err != nil {
			return fmt.Errorf("failed to read name %d length: %w", i, err)
		}
		nameLen := int(binary.BigEndian.Uint32(nameLenBuf))

		// Read string bytes
		nameBuf := make([]byte, nameLen)
		if _, err := io.ReadFull(file, nameBuf); err != nil {
			return fmt.Errorf("failed to read name %d: %w", i, err)
		}
		idx.names[i] = string(nameBuf)
	}

	return nil
}
