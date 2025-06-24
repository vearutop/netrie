package netrie

import (
	"net"
	"sort"
)

// ipToInt converts an IPv4 address to a uint32 for numerical comparison
func ipToInt(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

// intToIP converts a uint32 back to an IPv4 address
func intToIP(n uint32) net.IP {
	return net.IPv4(
		byte(n>>24),
		byte(n>>16),
		byte(n>>8),
		byte(n),
	)
}

// findCIDR finds the smallest CIDR that contains the range [start, end]
func findCIDR(start, end uint32) *net.IPNet {
	var mask uint32 = 0xffffffff
	bits := 32
	for bits > 0 {
		// Check if the range fits within the current mask
		if start&mask == end&mask {
			// Construct the CIDR
			ip := intToIP(start)
			return &net.IPNet{
				IP:   ip,
				Mask: net.CIDRMask(bits, 32),
			}
		}
		bits--
		mask <<= 1
	}
	// If no common prefix, return single IP as /32
	ip := intToIP(start)
	return &net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(32, 32),
	}
}

// ClusterIPs groups a sorted list of IPs into CIDR blocks
func ClusterIPs(ipStrings []string) []*net.IPNet {
	// Parse and convert IPs to uint32
	ips := make([]uint32, 0, len(ipStrings))
	for _, ipStr := range ipStrings {
		ip := net.ParseIP(ipStr)
		if ip == nil || ip.To4() == nil {
			continue // Skip invalid or non-IPv4 addresses
		}
		ips = append(ips, ipToInt(ip))
	}

	// Sort IPs numerically
	sort.Slice(ips, func(i, j int) bool { return ips[i] < ips[j] })

	// Cluster into CIDRs
	var cidrs []*net.IPNet
	start := 0
	for i := 1; i <= len(ips); i++ {
		// If we reach the end or find a gap, process the current range
		if i == len(ips) || ips[i] > ips[i-1]+1 {
			cidr := findCIDR(ips[start], ips[i-1])
			cidrs = append(cidrs, cidr)
			start = i
		}
	}

	return cidrs
}

// ipRange represents a range of IPs [start, end]
type ipRange struct {
	start, end uint32
}

// cidrToRange converts a CIDR to an ipRange (start and end IPs)
func cidrToRange(cidr *net.IPNet) ipRange {
	start := ipToInt(cidr.IP)
	mask := ipToInt(net.IP(cidr.Mask))
	end := start | ^mask // Last IP in the range
	return ipRange{start, end}
}

// rangeToCIDRs converts an ipRange to one or more CIDRs
func rangeToCIDRs(r ipRange) []*net.IPNet {
	var cidrs []*net.IPNet
	start, end := r.start, r.end

	for start <= end {
		// Find the largest CIDR that starts at 'start' and doesn't exceed 'end'
		var bits int
		var mask uint32 = 0xffffffff
		for bits = 32; bits >= 0; bits-- {
			if start&(1<<(32-bits)) != 0 {
				break // Can't use this bit size if start has a 1 in this position
			}
			networkSize := uint32(1 << (32 - bits))
			if start+networkSize-1 > end || (start&^mask) != 0 {
				bits++ // Step back if the network is too large or doesn't align
				break
			}
			mask <<= 1
		}
		if bits > 32 {
			bits = 32
		}

		// Create CIDR
		ip := intToIP(start)
		cidr := &net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(bits, 32),
		}
		cidrs = append(cidrs, cidr)

		// Move to the next address after this CIDR
		start += 1 << (32 - bits)
	}

	return cidrs
}

// MergeCIDRs merges a list of CIDRs into the smallest possible set
func MergeCIDRs(cidrStrings []string) []*net.IPNet {
	// Parse CIDRs and convert to ranges
	var ranges []ipRange
	for _, cidrStr := range cidrStrings {
		_, cidr, err := net.ParseCIDR(cidrStr)
		if err != nil {
			continue // Skip invalid CIDRs
		}
		if cidr.IP.To4() == nil {
			continue // Skip non-IPv4 CIDRs
		}
		ranges = append(ranges, cidrToRange(cidr))
	}

	// Sort ranges by start address
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].start < ranges[j].start || (ranges[i].start == ranges[j].start && ranges[i].end < ranges[j].end)
	})

	// Merge overlapping or adjacent ranges
	var merged []ipRange
	if len(ranges) == 0 {
		return nil
	}
	current := ranges[0]
	for i := 1; i < len(ranges); i++ {
		if ranges[i].start <= current.end+1 {
			// Overlapping or adjacent, extend current range
			if ranges[i].end > current.end {
				current.end = ranges[i].end
			}
		} else {
			// Non-adjacent, save current and start new
			merged = append(merged, current)
			current = ranges[i]
		}
	}
	merged = append(merged, current)

	// Convert merged ranges to CIDRs
	var result []*net.IPNet
	for _, r := range merged {
		result = append(result, rangeToCIDRs(r)...)
	}

	return result
}
