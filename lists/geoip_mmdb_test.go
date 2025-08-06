package lists_test

import (
	"net"
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
	"github.com/stretchr/testify/require"
	"github.com/vearutop/netrie"
	"github.com/vearutop/netrie/lists"
)

func TestLoadCityMMDB(t *testing.T) {
	tr := netrie.NewCIDRIndex()
	mmdbPath := path.Join(os.TempDir() + "GeoIP2-City.mmdb")

	if err := lists.GeoIPCityMMDB(tr, mmdbPath, nil); err != nil {
		t.Fatal(err)
	}

	println("nets", tr.Len())
	println("names", tr.LenNames())

	tr.SaveToFile("cities.bin")

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	println("HEAP:", ms.HeapAlloc)

	println(tr.Lookup("172.224.227.36"))
	println(tr.Lookup("178.15.138.158"))
	println(tr.Lookup("66.249.66.71"))
	println(tr.Lookup("143.198.196.44"))
}

func TestLoadCitiesFile(t *testing.T) {
	tr := netrie.NewCIDRIndex()

	err := tr.LoadFromFile("cities.bin")
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

func BenchmarkCitiesNetrieLookup(b *testing.B) {
	tr := netrie.NewCIDRIndex()

	require.NoError(b, tr.LoadFromFile("cities.bin"))

	println(tr.Len())

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	println("HEAP:", ms.HeapAlloc)

	b.ResetTimer()
	b.ReportAllocs()
	v := ""
	for i := 0; i < b.N; i++ {
		v = tr.Lookup("172.224.227.36")
	}

	println(v)
}

func BenchmarkMMDBLookup(b *testing.B) {
	st := time.Now()
	mmdbPath := path.Join(os.TempDir() + "GeoIP2-City.mmdb")
	println(mmdbPath)

	// db, err := geoip2.Open(mmdbPath)

	db, err := maxminddb.Open(mmdbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()
	println("open db:", time.Since(st).String())

	// If you are using strings that may be invalid, check that ip is not nil
	ip := net.ParseIP("172.224.227.36")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		city := geoip2.City{}
		_, _, err := db.LookupNetwork(ip, &city)
		//_, err := db.City(ip)
		if err != nil {
			b.Fatal(err)
		}

	}
}
