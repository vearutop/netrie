package netrie

import (
	"fmt"
	"github.com/thcyron/cidrmerge"
	"net"
)

// ClusterCIDRs takes IPs or CIDRs and returns minimal set of aggregated CIDRs.
func ClusterCIDRs2(input []string) ([]*net.IPNet, error) {
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
