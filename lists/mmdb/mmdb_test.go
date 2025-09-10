package mmdb_test

import (
	"github.com/stretchr/testify/require"
	"net"
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
	"github.com/vearutop/netrie"
	"github.com/vearutop/netrie/lists/mmdb"
)

func TestLoadCityMMDB2(t *testing.T) {
	tr := netrie.NewCIDRIndex[int32]()
	mmdbPath := path.Join("./GeoIP2-City.mmdb")

	if err := mmdb.LoadCityMMDB(tr, mmdbPath, nil); err != nil {
		t.Fatal(err)
	}

	println("nets", tr.Len())
	println("names", tr.LenNames())

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	println("HEAP:", ms.HeapAlloc)

	println(tr.Lookup("172.224.227.36")) // Coventry:GB:52.40640:-1.50820
	println(tr.Lookup("178.15.138.158")) // Berlin:DE:52.55300:13.45280
	println(tr.Lookup("66.249.66.71"))   // Foley:US:37.75100:-97.82200
	println(tr.Lookup("143.198.196.44")) // Singapore:SG:1.31400:103.68390
}

func TestLoadCityMMDB(t *testing.T) {
	tr := netrie.NewCIDRIndex[int32]()
	//mmdbPath := path.Join(os.TempDir() + "GeoIP2-City.mmdb")
	mmdbPath := path.Join("./GeoIP2-City.mmdb")

	if err := mmdb.LoadCityMMDB(tr, mmdbPath, nil); err != nil {
		t.Fatal(err)
	}

	println("nets", tr.Len())
	println("names", tr.LenNames())

	require.NoError(t, tr.SaveToFile("cities.bin"))

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)
	println("HEAP:", ms.HeapAlloc)

	println(tr.Lookup("172.224.227.36"))
	println(tr.Lookup("178.15.138.158"))
	println(tr.Lookup("66.249.66.71"))
	println(tr.Lookup("143.198.196.44"))
}

func TestLoadDisposableCloud(t *testing.T) {
	tr := netrie.NewCIDRIndex[int32]()

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

func BenchmarkMMDBLookupAny(b *testing.B) {
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
		var city map[string]any
		_, _, err := db.LookupNetwork(ip, &city)
		//_, err := db.City(ip)
		if err != nil {
			b.Fatal(err)
		}

	}
}
