package netrie

import (
	"bufio"
	"encoding/binary"
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

func (idx *CIDRIndex[S]) Save(w io.Writer) error {
	// Write header: total (int32), nodesLen (int32), namesLen (int32)
	header := make([]byte, 12)
	binary.BigEndian.PutUint32(header[0:4], uint32(idx.total))
	binary.BigEndian.PutUint32(header[4:8], uint32(len(idx.nodes)))
	binary.BigEndian.PutUint32(header[8:12], uint32(len(idx.names)))
	if _, err := w.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	var s S

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

func (idx *CIDRIndex[S]) Load(r io.Reader) error {
	// Read header: total (int32), nodesLen (int32), namesLen (int32)
	header := make([]byte, 12)
	if _, err := io.ReadFull(r, header); err != nil {
		return fmt.Errorf("read header: %w", err)
	}
	total := int(binary.BigEndian.Uint32(header[0:4]))
	nodesLen := int(binary.BigEndian.Uint32(header[4:8]))
	namesLen := int(binary.BigEndian.Uint32(header[8:12]))

	// Initialize CIDRIndex fields
	idx.total = total
	idx.nodes = make([]trieNode[S], nodesLen)
	idx.names = make([]string, namesLen)

	var s S

	switch any(s).(type) {
	case int16:
		// Read nodes
		nodeBuf := make([]byte, 11)
		for i := 0; i < nodesLen; i++ {
			if _, err := io.ReadFull(r, nodeBuf); err != nil {
				return fmt.Errorf("read node %d: %w", i, err)
			}
			if err := idx.nodes[i].UnmarshalBinary(nodeBuf); err != nil {
				return fmt.Errorf("unmarshal node %d: %w", i, err)
			}
		}
	case int32:
		// Read nodes
		nodeBuf := make([]byte, 13)
		for i := 0; i < nodesLen; i++ {
			if _, err := io.ReadFull(r, nodeBuf); err != nil {
				return fmt.Errorf("read node %d: %w", i, err)
			}
			if err := idx.nodes[i].UnmarshalBinary(nodeBuf); err != nil {
				return fmt.Errorf("unmarshal node %d: %w", i, err)
			}
		}
	}

	// Read names
	for i := 0; i < namesLen; i++ {
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

// LoadFromFile loads the CIDRIndex from a file.
func (idx *CIDRIndex[S]) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	r := bufio.NewReader(file)

	return idx.Load(r)
}
