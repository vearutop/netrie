// Package mmdb provides MaxMind GeoIP MMDB importer.
package mmdb

import (
	"log"
	"strconv"

	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
	"github.com/vearutop/netrie"
)

func LoadCityMMDB(tr *netrie.CIDRIndex[int32], mmdbPath string, makeName func(city geoip2.City) string) error {
	if makeName == nil {
		makeName = func(city geoip2.City) string {
			return city.City.Names["en"] + ":" + city.Country.IsoCode + ":" +
				strconv.FormatFloat(city.Location.Latitude, 'f', 5, 64) + ":" +
				strconv.FormatFloat(city.Location.Longitude, 'f', 5, 64)
		}
	}

	db, err := maxminddb.Open(mmdbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = db.Close()
	}()

	// skip aliased networks
	networks := db.Networks(maxminddb.SkipAliasedNetworks)
	if networks.Err() != nil {
		return networks.Err()
	}

	i := 0

	// City is nested
	record := geoip2.City{}
	for networks.Next() {
		subnet, err := networks.Network(&record)
		if err != nil {
			return err
		}

		name := makeName(record)
		tr.AddNet(subnet, name)
		i++
	}

	return nil
}

// /home/backend_team/vearutop/flatjsonl -children-limit 1000 -geo-ip-db /usr/share/GeoIP/GeoIP2-City.mmdb -geo-ip-db /usr/share/GeoIP/GeoIP2-Connection-Type.mmdb -geo-ip-db /usr/share/GeoIP/GeoIP2-ISP.mmdb -config /home/am.emea/alex.tomenko/logs/test_aw_install.json5 -show-keys-info -input /home/am.emea/alex.tomenko/logs/saygames_ca17ki5xuhhc_aw_installs_sep_1_8.log
// /home/backend_team/vearutop/flatjsonl -geo-ip-db /usr/share/GeoIP/GeoIP2-City.mmdb -geo-ip-db /usr/share/GeoIP/GeoIP2-Connection-Type.mmdb -geo-ip-db /usr/share/GeoIP/GeoIP2-ISP.mmdb -config ~/alex.tomenko/logs/test_aw_install.json5 -show-keys-info -input ~/alex.tomenko/logs/saygames_ca17ki5xuhhc_aw_installs_sep_1_8.log

// /home/backend_team/vearutop/flatjsonl -children-limit 1000 -config '{"extractValuesRegex":{".context.session.CallbackData.IpAddress":"GEOIP"}}' -geo-ip-db /usr/share/GeoIP/GeoIP2-City.mmdb -geo-ip-db /usr/share/GeoIP/GeoIP2-Connection-Type.mmdb -geo-ip-db /usr/share/GeoIP/GeoIP2-ISP.mmdb  -show-keys-info -input /home/am.emea/alex.tomenko/logs/saygames_ca17ki5xuhhc_aw_installs_sep_1_8.log | grep IpAddress
