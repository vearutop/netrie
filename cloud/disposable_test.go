package cloud_test

import (
	"net"
	"runtime"
	"testing"

	"github.com/vearutop/netrie"
	"github.com/vearutop/netrie/cloud"
)

func TestLoadDisposableCloudRanges(t *testing.T) {
	tr := netrie.NewCIDRIndex()

	err := cloud.LoadDisposableCloudRanges(tr)
	if err != nil {
		t.Fatal(err)
	}

	println(tr.Len())

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	tr.SaveToFile("ipranges.bin")

	println("HEAP:", ms.HeapAlloc)

	println("should be apple", tr.Lookup("172.224.227.36"))
	println("should be empty", tr.Lookup("178.15.138.158"))
}

func TestLoadDisposableCloud(t *testing.T) {
	tr := netrie.NewCIDRIndex()

	err := tr.LoadFromFile("ipranges.bin")
	if err != nil {
		t.Fatal(err)
	}

	println(tr.Len())

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	println("HEAP:", ms.HeapAlloc)

	println(tr.Lookup("172.224.227.36"))
	println(tr.Lookup("178.15.138.158"))
	println(tr.Lookup("66.249.66.71"))
	println(tr.Lookup("143.198.196.44"))
}

func BenchmarkLoadDisposableCloudRanges(b *testing.B) {
	tr := netrie.NewCIDRIndex()

	err := tr.LoadFromFile("ipranges.bin")
	if err != nil {
		b.Fatal(err)
	}

	println(tr.Len(), "nets")

	b.ResetTimer()
	b.ReportAllocs()

	ip := net.ParseIP("172.224.227.36").To4()

	for i := 0; i < b.N; i++ {
		_ = tr.LookupIP(ip)
	}

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	println("HEAP:", ms.HeapAlloc)

	println(tr.Lookup("172.224.227.36"))
}
