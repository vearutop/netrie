package lists

import (
	"log"
	"strconv"

	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
	"github.com/vearutop/netrie"
)

func GeoIPCityMMDB(tr *netrie.CIDRIndex, mmdbPath string, makeName func(city geoip2.City) string) error {
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

	record := geoip2.City{}
	for networks.Next() {
		subnet, err := networks.Network(&record)
		if err != nil {
			return err
		}

		name := makeName(record)
		tr.AddNet(subnet, name)
	}

	return nil
}
