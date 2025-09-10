package lists_test

import (
	"runtime"
	"testing"

	"github.com/vearutop/netrie"
	"github.com/vearutop/netrie/lists"
)

func TestLoadTorExitNodes(t *testing.T) {
	tr := netrie.NewCIDRIndex()

	err := lists.LoadTorExitNodes(tr)
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
}
