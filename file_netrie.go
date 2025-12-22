package netrie

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

// CIDRIndexFile is the trie structure for CIDR lookups.
type CIDRIndexFile[S int16 | int32] struct {
	meta Metadata

	mu sync.Mutex
	r  io.ReaderAt

	nodesOffset int64
	nodeSize    int64
	nodesLen    int64

	namesOffset int64
	namesLen    int64

	names []string
	total int
}

func newCIDRIndexFile[S int16 | int32](r io.ReaderAt, h hd) (*CIDRIndexFile[S], error) {
	nodesOffset := 20 + int64(h.metadataLen)

	idx := &CIDRIndexFile[S]{}
	idx.r = r
	idx.nodeSize = h.nodeSize
	idx.nodesLen = int64(h.nodesLen)
	idx.nodesOffset = nodesOffset
	idx.namesOffset = nodesOffset + int64(h.nodesLen)*idx.nodeSize
	idx.namesLen = int64(h.namesLen)
	idx.meta = h.meta
	idx.total = int(h.total)

	if err := idx.readNames(); err != nil {
		return nil, err
	}

	return idx, nil
}

// Metadata returns a reference to the Metadata object associated with the CIDRIndex.
func (idx *CIDRIndexFile[S]) Metadata() *Metadata {
	return &idx.meta
}

// Len returns the number of CIDRs in the trie.
func (idx *CIDRIndexFile[S]) Len() int {
	return idx.total
}

// LenNames returns the number of different names in the trie.
func (idx *CIDRIndexFile[S]) LenNames() int {
	return int(idx.namesLen)
}

// Lookup finds the id of the CIDR that contains the given IP string.
// Returns "" if no matching CIDR is found or IP is invalid.
func (idx *CIDRIndexFile[S]) Lookup(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "" // Invalid IP address.
	}

	return idx.LookupIP(ip)
}

func (idx *CIDRIndexFile[S]) readNames() error {
	offset := idx.namesOffset
	idx.names = make([]string, idx.namesLen)

	// Read names
	for i := 0; i < int(idx.namesLen); i++ {
		// Read string length (int32)
		nameLenBuf := make([]byte, 4)
		_, err := idx.r.ReadAt(nameLenBuf, offset)
		if err != nil && err != io.EOF {
			return fmt.Errorf("read name %d length: %w", i, err)
		}

		nameLen := int(binary.BigEndian.Uint32(nameLenBuf))
		offset += 4

		// Read string bytes
		nameBuf := make([]byte, nameLen)
		if _, err := idx.r.ReadAt(nameBuf, offset); err != nil {
			return fmt.Errorf("read name %d len %d: %w", i, nameLen, err)
		}
		name := string(nameBuf)
		offset += int64(nameLen)
		idx.names[i] = name
	}

	return nil
}

func (idx *CIDRIndexFile[S]) readNode(r io.ReaderAt, id int64, b []byte) (trieNode[S], error) {
	n, err := r.ReadAt(b, idx.nodesOffset+id*idx.nodeSize)
	if err != nil {
		return trieNode[S]{}, fmt.Errorf("read node %d: %w", id, err)
	}

	if int64(n) != idx.nodeSize {
		return trieNode[S]{}, fmt.Errorf("read node %d: unexpected size %d", id, n)
	}

	var node trieNode[S]
	if err := node.UnmarshalBinary(b); err != nil {
		return trieNode[S]{}, fmt.Errorf("unmarshal node %d: %w", id, err)
	}

	return node, nil
}

// LookupIP finds the id of the CIDR that contains the given IP.
// Returns "" if no matching CIDR is found.
func (idx *CIDRIndexFile[S]) lookupIP(ip net.IP) (string, error) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

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

	// b := idx.pool.Get().(*[]byte)
	// defer idx.pool.Put(b)

	b := make([]byte, idx.nodeSize)

	for i := 0; i < maxBits; i++ {
		curNode, err := idx.readNode(idx.r, int64(current), b)
		if err != nil {
			return "", err
		}

		// Check if current node has an id and update best match if mask is longer.
		if curNode.id != -1 && curNode.maskLen > bestMaskLen {
			bestID = curNode.id
			bestMaskLen = curNode.maskLen
		}

		// Get the next bit.
		bit := (ip[i/8] >> (7 - (i % 8))) & 1
		childIndex := curNode.children[bit]
		if childIndex == -1 {
			break // No further path.
		}
		current = int(childIndex)
	}

	curNode, err := idx.readNode(idx.r, int64(current), b)
	if err != nil {
		return "", err
	}

	// Check the final node for a better match.
	if curNode.id != -1 && curNode.maskLen > bestMaskLen {
		bestID = curNode.id
	}

	if bestID == -1 {
		return "", nil
	}

	return idx.names[bestID-1], nil
}

// SafeLookupIP performs a secure lookup for the given IP within the CIDRIndexFile.
// Returns the associated name and an error if the lookup fails.
func (idx *CIDRIndexFile[S]) SafeLookupIP(ip net.IP) (string, error) {
	return idx.lookupIP(ip)
}

// LookupIP finds the name associated with the CIDR containing the given IP.
// Panics if an internal error occurs during the lookup.
// Returns an empty string if no matching CIDR is found.
func (idx *CIDRIndexFile[S]) LookupIP(ip net.IP) string {
	name, err := idx.lookupIP(ip)
	if err != nil {
		panic(err)
	}

	return name
}

// Close releases any resources associated with the CIDRIndexFile, calling Close on the underlying io.Closer if available.
func (idx *CIDRIndexFile[S]) Close() error {
	if c, ok := idx.r.(io.Closer); ok {
		return c.Close()
	}

	return nil
}

// Options represents configuration options for customizing behaviors, such as buffer size for data readers.
type Options struct {
	BufferSize int // Default 4096.
}

// OpenFile opens a file at the specified path and parses it into a SafeIPLookuper
// for IP lookups with optional configurations.
func OpenFile(fn string, opts ...func(o *Options)) (SafeIPLookuper, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}

	return Open(f, opts...)
}

// Open parses a ReaderAt to load an SafeIPLookuper instance for performing IP lookups from a CIDR-based structure.
// Returns the constructed SafeIPLookuper and an error if any issue occurs during parsing or initialization.
func Open(r io.ReaderAt, opts ...func(o *Options)) (SafeIPLookuper, error) {
	o := Options{}
	o.BufferSize = 4096

	for _, opt := range opts {
		opt(&o)
	}

	if o.BufferSize > 0 {
		r = newBufReaderAt(r, o.BufferSize)
	}

	// Read header: version (uint32), total (uint32), nodesLen (uint32), namesLen (uint32), metadataLen (uint32)
	header := make([]byte, 20)

	if _, err := r.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	h := hd{}
	if err := h.UnmarshalBinary(header); err != nil {
		return nil, fmt.Errorf("unmarshal header: %w", err)
	}

	if h.metadataLen > 0 {
		metadataBuf := make([]byte, h.metadataLen)

		if _, err := r.ReadAt(metadataBuf, 20); err != nil {
			return nil, fmt.Errorf("read metadata: %w", err)
		}

		if err := json.Unmarshal(metadataBuf, &h.meta); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	if h.hasLargeNamespace {
		return newCIDRIndexFile[int32](r, h)
	}

	return newCIDRIndexFile[int16](r, h)
}

// bufReaderAt implements buffering for an io.ReaderAt object.
type bufReaderAt struct {
	offset     int64
	buf        []byte
	readerAt   io.ReaderAt
	sizeTilEOF int
	lastErr    error
}

func newBufReaderAt(readerAt io.ReaderAt, size int) *bufReaderAt {
	return &bufReaderAt{
		offset:     0,
		readerAt:   readerAt,
		buf:        make([]byte, size),
		sizeTilEOF: -1,
	}
}

func (r *bufReaderAt) ReadAt(b []byte, offset int64) (n int, err error) {
	bufEnd := offset + int64(len(b))

	switch {
	case (offset < r.cacheOffset(offset) && r.cacheOffset(offset) < bufEnd) ||
		(offset < r.expectedCacheEnd(offset) && r.expectedCacheEnd(offset) < bufEnd):

		return r.readerAt.ReadAt(b, offset)
	case !r.isInitialized() || r.expectedCacheEnd(offset) <= r.offset || r.cacheEnd() <= r.cacheOffset(offset):
		if n, err = r.readAtAndRenewCache(offset); n == 0 {
			return n, err
		}

		return r.copySafe(b, offset)
	case r.offset <= offset && bufEnd <= r.cacheEnd():

		return r.copySafe(b, offset)
	}

	return n, err
}

func (r *bufReaderAt) bufSize() int64 {
	return int64(len(r.buf))
}

func (r *bufReaderAt) cacheOffset(offset int64) (i int64) {
	return (offset / r.bufSize()) * r.bufSize()
}

func (r *bufReaderAt) expectedCacheEnd(offset int64) (i int64) {
	return r.cacheOffset(offset) + r.bufSize()
}

func (r *bufReaderAt) cacheEnd() int64 {
	return r.offset + r.bufSize()
}

func (r *bufReaderAt) readAtAndRenewCache(offset int64) (n int, err error) {
	expOffset := r.cacheOffset(offset)

	if r.isInitialized() && expOffset == r.offset {
		return 0, nil
	}

	// cache miss
	r.reset(expOffset)
	n, err = r.readerAt.ReadAt(r.buf, expOffset)

	r.sizeTilEOF = n
	r.lastErr = err

	return n, err
}

func (r *bufReaderAt) copySafe(b []byte, off int64) (n int, err error) {
	bufEnd := off + int64(len(b))
	s := int(off - r.cacheOffset(off))

	var e int
	switch {
	case r.dataEnd() <= off:
		return 0, io.EOF
	case off < r.dataEnd() && r.dataEnd() < bufEnd:
		e = r.sizeTilEOF
		err = r.lastErr
	case bufEnd <= r.dataEnd():
		e = s + len(b)
	}

	copy(b, r.buf[s:e])

	return e - s, err
}

func (r *bufReaderAt) reset(offset int64) {
	r.offset = offset
	r.sizeTilEOF = -1
	r.lastErr = nil
}

func (r *bufReaderAt) dataEnd() int64 {
	return r.offset + int64(r.sizeTilEOF)
}

func (r *bufReaderAt) isInitialized() bool {
	return 0 <= r.sizeTilEOF
}
