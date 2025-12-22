package mmdb_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vearutop/netrie"
	"github.com/vearutop/netrie/mmdb"
)

func TestLoadMMDB_city(t *testing.T) {
	tr := netrie.NewCIDRIndex()

	assertTr := func(t *testing.T, tr netrie.IPLookuper) {
		t.Helper()
		assert.Equal(t, 250, tr.Len())
		assert.Equal(t, 55, tr.LenNames())
		assert.Equal(t, "GB:Boxford", tr.Lookup("2.125.160.217"))
		assert.Equal(t, "GB:London", tr.Lookup("81.2.69.145"))
		assert.Equal(t, "US:San Diego", tr.Lookup("2001:480:10::1"))
		assert.Equal(t, "", tr.Lookup("143.198.196.44"))
		assert.Equal(t, "2025-08-12 17:49:01 +0000 UTC", tr.Metadata().BuildDate.String())
	}

	require.NoError(t, mmdb.Load(tr, "testdata/GeoIP2-City-Test.mmdb", mmdb.CityCountryISOCode))
	assertTr(t, tr)

	assert.Equal(t, 1486, tr.LenNodes())

	tr.Minimize()
	assertTr(t, tr)

	assert.Equal(t, 769, tr.LenNodes())

	require.NoError(t, tr.SaveToFile("testdata/cities.bin"))

	tr2, err := netrie.LoadFromFile("testdata/cities.bin")
	require.NoError(t, err)
	assertTr(t, tr2)

	f, err := os.Open("testdata/cities.bin")
	require.NoError(t, err)
	defer f.Close()

	tr3, err := netrie.Open(f)
	require.NoError(t, err)
	assertTr(t, tr3)
}

func TestLoadMMDB_country(t *testing.T) {
	tr := netrie.NewCIDRIndex()

	assertTr := func(t *testing.T, tr netrie.IPLookuper) {
		t.Helper()
		assert.Equal(t, 244, tr.Len())
		assert.Equal(t, 47, tr.LenNames())
		assert.Equal(t, "GB", tr.Lookup("2.125.160.217"))
		assert.Equal(t, "GB", tr.Lookup("81.2.69.145"))
		assert.Equal(t, "US", tr.Lookup("2001:480:10::1"))
		assert.Equal(t, "", tr.Lookup("143.198.196.44"))
		assert.Equal(t, "2025-08-12 17:49:01 +0000 UTC", tr.Metadata().BuildDate.String())
	}

	require.NoError(t, mmdb.Load(tr, "testdata/GeoLite2-Country-Test.mmdb", mmdb.CountryISOCode))
	assertTr(t, tr)

	assert.Equal(t, 1434, tr.LenNodes())

	tr.Minimize()
	assertTr(t, tr)

	assert.Equal(t, 714, tr.LenNodes())

	require.NoError(t, tr.SaveToFile("testdata/countries.bin"))

	tr2, err := netrie.LoadFromFile("testdata/countries.bin")
	require.NoError(t, err)
	assertTr(t, tr2)

	f, err := os.Open("testdata/countries.bin")
	require.NoError(t, err)
	defer f.Close()

	tr3, err := netrie.Open(f)
	require.NoError(t, err)
	assertTr(t, tr3)
}

func TestLoadMMDB_asn(t *testing.T) {
	tr := netrie.NewCIDRIndex()
	tr.Metadata().Description = "ASNs"

	assertTr := func(t *testing.T, tr netrie.IPLookuper) {
		t.Helper()
		assert.Equal(t, 412, tr.Len())
		assert.Equal(t, 101, tr.LenNames())
		assert.Equal(t, "AS15169 Google Inc.", tr.Lookup("1.0.0.4"))
		assert.Equal(t, "AS701 MCI Communications Services, Inc. d/b/a Verizon Business", tr.Lookup("71.96.0.3"))
		assert.Equal(t, "AS1273 Cable and Wireless Worldwide plc", tr.Lookup("2001:5400::12"))
		assert.Equal(t, "", tr.Lookup("143.198.196.44"))
		assert.Equal(t, "2025-08-12 17:49:01 +0000 UTC", tr.Metadata().BuildDate.String())
		assert.Equal(t, "ASNs", tr.Metadata().Description)
	}

	require.NoError(t, mmdb.Load(tr, "testdata/GeoLite2-ASN-Test.mmdb", mmdb.ASNOrg))
	assertTr(t, tr)

	assert.Equal(t, 1434, tr.LenNodes())

	tr.Minimize()
	assertTr(t, tr)

	assert.Equal(t, 1208, tr.LenNodes())

	require.NoError(t, tr.SaveToFile("testdata/asns.bin"))

	tr2, err := netrie.LoadFromFile("testdata/asns.bin")
	require.NoError(t, err)
	assertTr(t, tr2)

	f, err := os.Open("testdata/asns.bin")
	require.NoError(t, err)
	defer f.Close()

	tr3, err := netrie.Open(f)
	require.NoError(t, err)
	assertTr(t, tr3)
}
