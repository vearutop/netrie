package main

import (
	"fmt"
	"net"
	"sort"
)

// parseIPOrCIDR parses an IP or CIDR into a *net.IPNet.
func parseIPOrCIDR(s string) (*net.IPNet, error) {
	if ip := net.ParseIP(s); ip != nil {
		maskLen := 32
		if ip.To4() == nil {
			maskLen = 128
		}
		return &net.IPNet{IP: ip, Mask: net.CIDRMask(maskLen, maskLen)}, nil
	}
	ip, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	ipNet.IP = ip
	return ipNet, nil
}

// ClusterCIDRs takes IPs or CIDRs and returns minimal set of aggregated CIDRs.
func ClusterCIDRs(input []string) ([]*net.IPNet, error) {
	var ipv4Nets []*net.IPNet
	var ipv6Nets []*net.IPNet

	for _, s := range input {
		ipNet, err := parseIPOrCIDR(s)
		if err != nil {
			return nil, fmt.Errorf("parse error for %q: %v", s, err)
		}
		if ipNet.IP.To4() != nil {
			ipv4Nets = append(ipv4Nets, ipNet)
		} else {
			ipv6Nets = append(ipv6Nets, ipNet)
		}
	}

	aggV4 := aggregateCIDRs(ipv4Nets)
	aggV6 := aggregateCIDRs(ipv6Nets)

	return append(aggV4, aggV6...), nil
}

// aggregateCIDRs merges adjacent or overlapping CIDRs.
func aggregateCIDRs(nets []*net.IPNet) []*net.IPNet {
	if len(nets) == 0 {
		return nil
	}

	// Sort by IP.
	sort.Slice(nets, func(i, j int) bool {
		return compareIPs(nets[i].IP, nets[j].IP) < 0
	})

	var result []*net.IPNet

	for _, net := range nets {
		result = appendAndMerge(result, net)
	}

	return result
}

// compareIPs compares two IP addresses lexicographically.
func compareIPs(a, b net.IP) int {
	a = a.To16()
	b = b.To16()
	return bytesCompare(a, b)
}

func bytesCompare(a, b []byte) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	return len(a) - len(b)
}

// appendAndMerge tries to merge a CIDR into the result slice.
func appendAndMerge(result []*net.IPNet, newNet *net.IPNet) []*net.IPNet {
	for {
		merged := false
		for i, r := range result {
			if cidrsCanMerge(r, newNet) {
				newNet = mergeCIDRs(r, newNet)
				result = append(result[:i], result[i+1:]...)
				merged = true
				break
			}
		}
		if !merged {
			break
		}
	}
	return append(result, newNet)
}

// cidrsCanMerge checks if two CIDRs can be merged into a larger one.
func cidrsCanMerge(a, b *net.IPNet) bool {
	onesA, bitsA := a.Mask.Size()
	onesB, bitsB := b.Mask.Size()

	if bitsA != bitsB || onesA != onesB || onesA == 0 {
		return false
	}

	supernet := &net.IPNet{
		IP:   a.IP.Mask(net.CIDRMask(onesA-1, bitsA)),
		Mask: net.CIDRMask(onesA-1, bitsA),
	}

	return supernet.Contains(a.IP) && supernet.Contains(b.IP)
}

// mergeCIDRs returns the merged CIDR of two adjacent networks.
func mergeCIDRs(a, b *net.IPNet) *net.IPNet {
	ones, bits := a.Mask.Size()
	return &net.IPNet{
		IP:   a.IP.Mask(net.CIDRMask(ones-1, bits)),
		Mask: net.CIDRMask(ones-1, bits),
	}
}
