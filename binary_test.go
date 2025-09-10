package netrie

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrieNode_MarshalBinary(t *testing.T) {
	tn := trieNode{
		children: [2]int32{5678, 1234},
		id:       3456,
		maskLen:  68,
	}

	b, err := tn.MarshalBinary()
	require.NoError(t, err)
	assert.Len(t, b, 11)

	tn2 := trieNode{}
	require.NoError(t, tn2.UnmarshalBinary(b))

	assert.Equal(t, tn.children, tn2.children)
	assert.Equal(t, tn.id, tn2.id)
	assert.Equal(t, tn.maskLen, tn2.maskLen)
}

func TestCIDRIndex_Load(t *testing.T) {
	tr := NewCIDRIndex()

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
		require.NoError(t, tr.AddCIDR(c.cidr, c.name), c.cidr)
	}

	buf := bytes.NewBuffer(nil)
	require.NoError(t, tr.Save(buf))

	tr2 := NewCIDRIndex()
	require.NoError(t, tr2.Load(buf))

	assert.Equal(t, 4, tr2.Len())
	assert.Equal(t, 4, tr2.LenNames())

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
		name := tr2.Lookup(tc.ip)
		if name != tc.name {
			t.Errorf("Lookup(%q) = %s, want %s", tc.ip, name, tc.name)
		}
	}
}
