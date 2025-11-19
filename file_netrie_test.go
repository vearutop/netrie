package netrie_test

import (
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vearutop/netrie"
)

func TestLoadFromFile(t *testing.T) {
	assertTr := func(t *testing.T, tr netrie.IPLookuper) {
		t.Helper()

		wg := sync.WaitGroup{}
		wg.Add(100)

		for i := 0; i < 100; i++ {
			go func() {
				defer wg.Done()
				assert.Equal(t, 250, tr.Len())
				assert.Equal(t, 55, tr.LenNames())
				assert.Equal(t, "GB:Boxford", tr.Lookup("2.125.160.217"))
				assert.Equal(t, "GB:London", tr.Lookup("81.2.69.145"))
				assert.Equal(t, "US:San Diego", tr.Lookup("2001:480:10::1"))
				assert.Equal(t, "", tr.Lookup("143.198.196.44"))
				assert.Equal(t, "2025-08-12 17:49:01 +0000 UTC", tr.Metadata().BuildDate.String())
			}()
		}

		wg.Wait()
	}

	tr2, err := netrie.LoadFromFile("testdata/cities.bin")
	require.NoError(t, err)
	assertTr(t, tr2)

	f, err := os.Open("testdata/cities.bin")
	require.NoError(t, err)
	defer f.Close()

	tr3, err := netrie.Open(f, func(o *netrie.Options) {
		o.BufferSize = 0
	})
	require.NoError(t, err)
	assertTr(t, tr3)

	tr4, err := netrie.Open(f)
	require.NoError(t, err)
	assertTr(t, tr4)
}

func BenchmarkLoadMMDB_city(b *testing.B) {
	assertTr := func(b *testing.B, tr netrie.IPLookuper) {
		b.Helper()
		assert.Equal(b, 250, tr.Len())
		assert.Equal(b, 55, tr.LenNames())
		assert.Equal(b, "GB:Boxford", tr.Lookup("2.125.160.217"))
		assert.Equal(b, "GB:London", tr.Lookup("81.2.69.145"))
		assert.Equal(b, "US:San Diego", tr.Lookup("2001:480:10::1"))
		assert.Equal(b, "", tr.Lookup("143.198.196.44"))
		assert.Equal(b, "2025-08-12 17:49:01 +0000 UTC", tr.Metadata().BuildDate.String())
	}

	tr2, err := netrie.LoadFromFile("testdata/cities.bin")
	require.NoError(b, err)

	f, err := os.Open("testdata/cities.bin")
	require.NoError(b, err)
	defer f.Close()

	tr3, err := netrie.Open(f, func(o *netrie.Options) {
		o.BufferSize = 0
	})
	require.NoError(b, err)

	tr4, err := netrie.Open(f)
	require.NoError(b, err)

	b.Run("mem", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			assertTr(b, tr2)
		}
	})

	b.Run("file", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			assertTr(b, tr3)
		}
	})

	b.Run("buf_file", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			assertTr(b, tr4)
		}
	})
}
