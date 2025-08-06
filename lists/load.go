package lists

import (
	"github.com/vearutop/netrie"
)

func loadFromText(u string, tr *netrie.CIDRIndex, name string) error {
	return netrie.LoadFromTextCB(u, func(s string) error {
		return tr.AddCIDR(s, name)
	})
}
