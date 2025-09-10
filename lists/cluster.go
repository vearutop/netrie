package lists

import (
	"fmt"
	"net"

	"github.com/thcyron/cidrmerge"
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
	var nets []*net.IPNet

	for _, s := range input {
		ipNet, err := parseIPOrCIDR(s)
		if err != nil {
			return nil, fmt.Errorf("parse error for %q: %v", s, err)
		}

		nets = append(nets, ipNet)
	}

	merged := cidrmerge.Merge(nets)

	return merged, nil
}
