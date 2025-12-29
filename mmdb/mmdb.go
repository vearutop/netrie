// Package mmdb provides MaxMind GeoIP MMDB importer.
package mmdb

import (
	"encoding/json"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/oschwald/maxminddb-golang"
	"github.com/thcyron/cidrmerge"
	"github.com/vearutop/netrie"
)

// AnonymousIP configures the Options object to extract and process anonymous IP-related attributes from MaxMind DB records.
func AnonymousIP(o *Options) {
	o.MakeValueName = func() (any, func() string) {
		var v any
		return &v, func() string {
			// {"is_anonymous":true,"is_anonymous_vpn":true,"is_hosting_provider":true,"is_public_proxy":true,"is_residential_proxy":true,"is_tor_exit_node":true}

			var res string

			for _, k := range []string{
				"is_anonymous", "is_anonymous_vpn", "is_hosting_provider", "is_public_proxy",
				"is_residential_proxy", "is_tor_exit_node",
			} {
				if b, _ := RetrieveValue(v, k).(bool); b {
					res += k + ":"
				}
			}

			if res == "" {
				return ""
			}

			return res[:len(res)-1]
		}
	}
}

// CountryISOCode configures the Options to extract the ISO country code from a MaxMind DB record.
// If no code is found, "??" is used.
func CountryISOCode(o *Options) {
	o.MakeValueName = func() (any, func() string) {
		var v any
		return &v, func() string {
			isoCode, _ := RetrieveValue(v, "country", "iso_code").(string)

			if isoCode == "" {
				isoCode, _ = RetrieveValue(v, "registered_country", "iso_code").(string)
			}

			if isoCode == "" {
				isoCode = "??"
			}

			return isoCode
		}
	}
}

// RetrieveValue traverses nested maps using the provided keys and returns the corresponding value or nil if not found.
func RetrieveValue(v any, keys ...string) any {
	for _, k := range keys {
		if m, ok := v.(map[string]any); ok {
			v = m[k]
		} else {
			return nil
		}
	}

	return v
}

// CityCountryISOCode configures the Options to generate value names in the format "ISOCode:CityName" from records.
func CityCountryISOCode(o *Options) {
	o.MakeValueName = func() (any, func() string) {
		var v any
		return &v, func() string {
			cityName, _ := RetrieveValue(v, "city", "names", "en").(string)
			if cityName == "" {
				cityName = "Unknown"
			}

			isoCode, _ := RetrieveValue(v, "country", "iso_code").(string)

			if isoCode == "" {
				isoCode, _ = RetrieveValue(v, "country", "iso_code").(string)
			}

			if isoCode == "" {
				isoCode = "??"
			}

			return isoCode + ":" + cityName
		}
	}
}

// ASNOrg configures the `Options` structure to create a value-name function for AS numbers and organizations.
// It formats the result as "AS<number> <organization>" if available, or just "AS<number>" otherwise.
func ASNOrg(o *Options) {
	o.MakeValueName = func() (any, func() string) {
		var v any
		return &v, func() string {
			asnOrg, _ := RetrieveValue(v, "autonomous_system_organization").(string)

			aa := RetrieveValue(v, "autonomous_system_number")
			asn, _ := aa.(uint64)

			res := "AS" + strconv.Itoa(int(asn))

			if asnOrg != "" {
				res += " " + asnOrg
			}

			return res
		}
	}
}

// Options defines configurations for processing data, including a custom value-name function and printing options.
type Options struct {
	MakeValueName func() (any, func() string)
	PrintNets     bool
	PrintProgress bool
}

// Load loads MaxMind DB (MMDB) data into a trie structure, processing networks grouped by unique value names.
// It utilizes a custom function to extract values and names from database records.
func Load(tr netrie.Adder, mmdbPath string, options ...func(o *Options)) error {
	o := &Options{}

	for _, opt := range options {
		opt(o)
	}

	if o.MakeValueName == nil {
		o.MakeValueName = func() (any, func() string) {
			var v any
			return &v, func() string {
				j, _ := json.Marshal(v)

				return string(j)
			}
		}
	}

	db, err := maxminddb.Open(mmdbPath)
	if err != nil {
		return err
	}

	defer func() {
		_ = db.Close() //
	}()

	meta := tr.Metadata()
	meta.BuildDate = time.Unix(int64(db.Metadata.BuildEpoch), 0).UTC()
	if meta.Description == "" {
		meta.Description = db.Metadata.Description["en"]
	}

	if meta.Name == "" {
		meta.Name = strings.TrimSuffix(db.Metadata.Description["en"], " database")
	}
	meta.Extra = db.Metadata

	// skip aliased networks
	networks := db.Networks(maxminddb.SkipAliasedNetworks)
	if networks.Err() != nil {
		return networks.Err()
	}

	var (
		i        = 0
		blocks   = 0
		prevName = ""
		nets     []*net.IPNet
	)

	for networks.Next() {
		rec, nameFn := o.MakeValueName()

		subnet, err := networks.Network(rec)
		if err != nil {
			return err
		}

		name := nameFn()

		if prevName == "" && len(nets) == 0 {
			prevName = name
		} else if prevName != name {
			blocks++

			merged := cidrmerge.Merge(nets)

			for _, n := range merged {
				if o.PrintNets {
					println(n.String(), prevName)
				}

				tr.AddNet(n, prevName)
			}

			nets = nets[:0]

			prevName = name
		}

		nets = append(nets, subnet)

		i++

		if o.PrintProgress && i%10000 == 0 {
			println("original nets:", i, "merged blocks", blocks)
		}
	}

	merged := cidrmerge.Merge(nets)
	for _, n := range merged {
		tr.AddNet(n, prevName)
	}

	return nil
}
