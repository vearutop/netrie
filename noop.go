package netrie

import (
	"net"
)

// Noop is a placeholder type that implements various methods with empty or no-op behavior.
type Noop struct{}

// AddNet is a no-op method that accepts an IP network and a name but performs no actions.
func (n Noop) AddNet(ipNet *net.IPNet, name string) {}

// AddCIDR is a no-op method that accepts a CIDR block and a name but performs no actions and always returns nil.
func (n Noop) AddCIDR(cidr string, name string) error {
	return nil
}

// SafeLookupIP performs a safe lookup for the given IP, returning an empty string and nil error in this no-op implementation.
func (n Noop) SafeLookupIP(ip net.IP) (string, error) { return "", nil }

// LookupIP is a no-op method that accepts an IP and always returns an empty string.
func (n Noop) LookupIP(ip net.IP) string { return "" }

// Lookup is a no-op method that takes an IP address as a string and always returns an empty string.
func (n Noop) Lookup(ipStr string) string { return "" }

// Len returns the length of the internal data, always 0 in this no-op implementation.
func (n Noop) Len() int { return 0 }

// LenNames returns the number of names stored, always 0 in this no-op implementation.
func (n Noop) LenNames() int { return 0 }

// Metadata returns a pointer to a Metadata object, always nil in this no-op implementation.
func (n Noop) Metadata() *Metadata { return nil }

// Close is a no-op method that performs no actions and always returns nil.
func (n Noop) Close() error { return nil }
