package lists_test

import (
	"runtime"
	"testing"

	"github.com/vearutop/netrie"
)

func TestLoadTorExitNodes(t *testing.T) {
	tr := netrie.NewCIDRIndex()

	err := netrie.LoadFromTextGroupIPs("testdata/torlist.txt", tr, "tor-exit-nodes")
	if err != nil {
		t.Fatal(err)
	}

	println("nets:", tr.Len())
	println("names:", tr.LenNames())

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	tr.SaveToFile("ipranges-tor.bin")

	println("HEAP:", ms.HeapAlloc)

	println("should be apple", tr.Lookup("172.224.227.36"))
	println("should be empty", tr.Lookup("178.15.138.158"))
	println("should be tor", tr.Lookup("2001:067c:0e60:0c0c:0192:0042:0116:0185"))
	println("should be tor", tr.Lookup("192.42.116.214"))
}

func TestLoadTorExitNodes_file(t *testing.T) {
	tr := netrie.NewCIDRIndex()

	err := tr.LoadFromFile("ipranges-tor.bin")
	if err != nil {
		t.Fatal(err)
	}

	println("nets:", tr.Len())
	println("names:", tr.LenNames())

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	println("HEAP:", ms.HeapAlloc)

	println(tr.Lookup("172.224.227.36"))
	println(tr.Lookup("178.15.138.158"))
	println(tr.Lookup("66.249.66.71"))
	println(tr.Lookup("143.198.196.44"))
	println("should be tor", tr.Lookup("2001:067c:0e60:0c0c:0192:0042:0116:0185"))
	println("should be tor", tr.Lookup("192.42.116.214"))

}
