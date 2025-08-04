package lists_test

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"testing"

	"github.com/yl2chen/cidranger"
)

func BenchmarkCidranger(b *testing.B) {
	ranger := cidranger.NewPCTrieRanger()

	if err := addAppleICloudPrivateRelay(ranger); err != nil {
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

func addAppleICloudPrivateRelay(tr cidranger.Ranger) error {
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
