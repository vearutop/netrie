package lists

import "github.com/vearutop/netrie"

// LoadTorExitNodes loads TOR exit nodes from https://www.dan.me.uk/torlist/?exit.
func LoadTorExitNodes(tr *netrie.CIDRIndex) error {
	// return loadFromTextGroupIPs("https://www.dan.me.uk/torlist/?exit", tr, "tor-exit-nodes")
	return loadFromTextGroupCIDRs("testdata/torlist.txt", tr, "tor-exit-nodes")
}
