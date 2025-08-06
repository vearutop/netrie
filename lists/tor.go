package lists

import "github.com/vearutop/netrie"

// TorExitNodes loads TOR exit nodes from https://www.dan.me.uk/torlist/?exit.
func TorExitNodes(tr *netrie.CIDRIndex) error {
	return netrie.LoadFromTextGroupIPs("https://www.dan.me.uk/torlist/?exit", tr, "tor-exit-nodes")
}
