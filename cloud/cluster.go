package cloud

import "github.com/vearutop/netrie"

func MakeAhrefsCIDRs() {
	var ips []string

	ips = append(ips, "127.0.0.40")

	loadFromJSON("https://api.ahrefs.com/v3/public/crawler-ips", func(path []string, value interface{}) error {
		if len(path) == 3 && path[2] == "ip_address" {
			ips = append(ips, value.(string))
		}

		return nil
	})

	nets := netrie.ClusterIPs(ips)

	for _, n := range nets {
		println(n.String())
	}
}

func MakeAppleCIDRs() {
	var cidrs []string

	loadFromTextCB("https://mask-api.icloud.com/egress-ip-ranges.csv", func(s string) error {
		cidrs = append(cidrs, s)

		return nil
	})

	println("before:", len(cidrs))

	nets := netrie.MergeCIDRs(cidrs)

	println("after:", len(nets))

	for _, n := range nets {
		println(n.String())
	}

}
