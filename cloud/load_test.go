package cloud_test

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"testing"

	"github.com/vearutop/netrie"
	"github.com/vearutop/netrie/cloud"
	"github.com/yl2chen/cidranger"
)

func TestAddAppleICloudPrivateRelay(t *testing.T) {
	tr := netrie.NewCIDRIndex()

	if err := cloud.AddAppleICloudPrivateRelay(tr, 13); err != nil {
		t.Errorf("Error adding AppleICloudPrivateRelay: %v\n", err)
	}

	println(tr.Len())

	println(tr.Lookup("172.224.227.36"))
	println(tr.Lookup("185.70.52.107"))
}

func TestAddAWS(t *testing.T) {
	tr := netrie.NewCIDRIndex()

	if err := cloud.AddAWS(tr, 12); err != nil {
		t.Errorf("Error adding AWS: %v\n", err)
	}

	println(tr.Len())
}

func BenchmarkAddAppleICloudPrivateRelay(b *testing.B) {
	tr := netrie.NewCIDRIndex()

	if err := cloud.AddAppleICloudPrivateRelay(tr, 13); err != nil {
		b.Errorf("Error adding AppleICloudPrivateRelay: %v\n", err)
	}

	println(tr.Len(), "nets")

	b.ResetTimer()
	b.ReportAllocs()

	ip := net.ParseIP("172.224.227.36").To4()

	var v int16
	for i := 0; i < b.N; i++ {
		v += tr.LookupIP(ip)
	}

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	println(ms.HeapAlloc)

	v += tr.Lookup("172.224.227.36")
	_ = v
}

func BenchmarkCidranger(b *testing.B) {
	ranger := cidranger.NewPCTrieRanger()

	if err := addAppleICloudPrivateRelay(ranger, 13); err != nil {
		b.Errorf("Error adding AppleICloudPrivateRelay: %v\n", err)
		return
	}

	b.ResetTimer()
	b.ReportAllocs()

	ip := net.ParseIP("172.224.227.36")

	var v int16
	for i := 0; i < b.N; i++ {
		c, err := ranger.Contains(ip)
		if err != nil {
			b.Error(err)
		}

		if c {
			v++
		}
	}

	ms := runtime.MemStats{}
	runtime.ReadMemStats(&ms)

	println(ms.HeapAlloc)

	ranger.Contains(ip)
}

func addAppleICloudPrivateRelay(tr cidranger.Ranger, id int16) error {
	// https://raw.githubusercontent.com/femueller/cloud-ip-ranges/refs/heads/master/apple-icloud-private-relay-ip-ranges.csv

	req, err := http.NewRequest(http.MethodGet,
		"https://raw.githubusercontent.com/femueller/cloud-ip-ranges/refs/heads/master/apple-icloud-private-relay-ip-ranges.csv",
		nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad HTTP status code: %d", resp.StatusCode)
	}

	csvReader := csv.NewReader(resp.Body)
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		_, ipNet, err := net.ParseCIDR(record[0])
		if err != nil {
			return fmt.Errorf("invalid CIDR: %v", record[0])
		}
		//
		//Skip IPv6.
		//if ipNet.IP.To4() == nil {
		//	continue
		//}

		if err := tr.Insert(cidranger.NewBasicRangerEntry(*ipNet)); err != nil {
			return err
		}
	}

	return nil
}

func TestAddFastly(t *testing.T) {
	tr := netrie.NewCIDRIndex()
	if err := cloud.AddFastly(tr, 13); err != nil {
		t.Errorf("Error adding Fastly: %v\n", err)
	}

	println(tr.Len())
}

func TestAddGitHub(t *testing.T) {
	tr := netrie.NewCIDRIndex()
	if err := cloud.AddGitHub(tr, 13); err != nil {
		t.Errorf("Error adding github: %v\n", err)
	}

	println(tr.Len())
}

func TestAddGoogleCloud(t *testing.T) {
	tr := netrie.NewCIDRIndex()
	if err := cloud.AddGoogleCloud(tr, 13); err != nil {
		t.Errorf("Error adding google cloud: %v\n", err)
	}

	println(tr.Len())
}

func TestAddLinode(t *testing.T) {
	tr := netrie.NewCIDRIndex()
	if err := cloud.AddLinode(tr, 13); err != nil {
		t.Errorf("Error adding: %v\n", err)
	}

	println(tr.Len())
}

func TestAddMicrosoftAzure(t *testing.T) {
	tr := netrie.NewCIDRIndex()
	if err := cloud.AddMicrosoftAzure(tr, 13); err != nil {
		t.Errorf("Error adding: %v\n", err)
	}

	println(tr.Len())
}

func TestAddOracleCloud(t *testing.T) {
	tr := netrie.NewCIDRIndex()
	if err := cloud.AddOracleCloud(tr, 13); err != nil {
		t.Errorf("Error adding: %v\n", err)
	}

	println(tr.Len())
}

func TestAddZscalerCloud(t *testing.T) {
	tr := netrie.NewCIDRIndex()
	if err := cloud.AddZscalerCloud(tr, 13); err != nil {
		t.Errorf("Error adding: %v\n", err)
	}

	println(tr.Len())
}

func TestAddAkamai(t *testing.T) {
	tr := netrie.NewCIDRIndex()
	if err := cloud.AddAkamai(tr, 13); err != nil {
		t.Errorf("Error adding: %v\n", err)
	}

	println(tr.Len())
}

func TestAddCloudflare(t *testing.T) {
	tr := netrie.NewCIDRIndex()
	if err := cloud.AddCloudflare(tr, 13); err != nil {
		t.Errorf("Error adding: %v\n", err)
	}

	println(tr.Len())
}

func TestAddDigitalOcean(t *testing.T) {
	tr := netrie.NewCIDRIndex()
	if err := cloud.AddDigitalOcean(tr, 13); err != nil {
		t.Errorf("Error adding: %v\n", err)
	}

	println(tr.Len())
}
