package cloud

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/vearutop/netrie"
)

func AddAppleICloudPrivateRelay(tr *netrie.CIDRIndex, id int16) error {
	return loadFromText(
		"https://raw.githubusercontent.com/femueller/cloud-ip-ranges/refs/heads/master/apple-icloud-private-relay-ip-ranges.csv",
		tr,
		id,
	)
}

func AddAkamai(tr *netrie.CIDRIndex, id int16) error {
	return loadFromText(
		"https://raw.githubusercontent.com/femueller/cloud-ip-ranges/refs/heads/master/akamai-v4-ip-ranges.txt",
		tr,
		id,
	)
}

func AddCloudflare(tr *netrie.CIDRIndex, id int16) error {
	return loadFromText(
		"https://www.cloudflare.com/ips-v4",
		tr,
		id,
	)
}

func AddDigitalOcean(tr *netrie.CIDRIndex, id int16) error {
	return loadFromText(
		"https://raw.githubusercontent.com/femueller/cloud-ip-ranges/refs/heads/master/digitalocean.csv",
		tr,
		id,
	)
}

func AddFastly(tr *netrie.CIDRIndex, id int16) error {
	return loadFromJSON("https://raw.githubusercontent.com/femueller/cloud-ip-ranges/refs/heads/master/fastly-ip-ranges.json",
		func(path []string, value interface{}) error {
			if len(path) == 2 && path[0] == "addresses" {
				return tr.AddCIDR(value.(string), id)
			}
			return nil
		})
}

func AddGoogleCloud(tr *netrie.CIDRIndex, id int16) error {
	return loadFromJSON(
		"https://www.gstatic.com/ipranges/cloud.json",
		func(path []string, value interface{}) error {
			if len(path) == 3 && path[2] == "ipv4Prefix" {
				if err := tr.AddCIDR(value.(string), id); err != nil {
					//if errors.Is(err, netrie.ErrOverlap) {
					//	return nil
					//}

					return err
				}
			}

			return nil
		})
}

func AddGitHub(tr *netrie.CIDRIndex, id int16) error {
	return loadFromJSONBruteForce(
		"https://raw.githubusercontent.com/femueller/cloud-ip-ranges/refs/heads/master/github-ip-ranges.json",
		tr,
		id,
	)
}

func AddMicrosoftAzure(tr *netrie.CIDRIndex, id int16) error {
	return loadFromJSONBruteForce(
		"https://raw.githubusercontent.com/femueller/cloud-ip-ranges/refs/heads/master/microsoft-azure-ip-ranges.json",
		tr,
		id,
	)
}

func AddOracleCloud(tr *netrie.CIDRIndex, id int16) error {
	return loadFromJSONBruteForce(
		"https://docs.oracle.com/en-us/iaas/tools/public_ip_ranges.json",
		tr,
		id,
	)
}

func AddZscalerCloud(tr *netrie.CIDRIndex, id int16) error {
	return loadFromJSONBruteForce(
		"https://raw.githubusercontent.com/femueller/cloud-ip-ranges/refs/heads/master/zscaler-cloud-ip-ranges.json",
		tr,
		id,
	)
}

func AddAWS(tr *netrie.CIDRIndex, id int16) error {
	return loadFromJSON("https://ip-ranges.amazonaws.com/ip-ranges.json",
		func(path []string, value interface{}) error {
			if len(path) == 3 && path[2] == "ip_prefix" {
				if err := tr.AddCIDR(value.(string), id); err != nil {
					if errors.Is(err, netrie.ErrOverlap) {
						return nil
					}
				}
			}

			return nil
		})
}

func AddLinode(tr *netrie.CIDRIndex, id int16) error {
	return loadFromText(
		"https://geoip.linode.com/",
		tr,
		id,
	)
}

func loadFromJSONBruteForce(u string, tr *netrie.CIDRIndex, id int16) error {
	return loadFromJSON(
		u,
		func(path []string, value interface{}) error {
			if s, ok := value.(string); ok {
				if err := tr.AddCIDR(s, id); err != nil {
					if errors.Is(err, netrie.ErrOverlap) {
						return nil
					}

					return err
				}
			}
			return nil
		})
}

func loadFromJSON(u string, cb func(path []string, value interface{}) error) error {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad HTTP status code: %d", resp.StatusCode)
	}

	if err := walkJSON(resp.Body, cb); err != nil {
		return err
	}

	return nil
}

func loadFromTextCB(u string, cb func(value string) error) error {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad HTTP status code: %d", resp.StatusCode)
	}

	r := bufio.NewScanner(resp.Body)

	for r.Scan() {
		line := strings.TrimSpace(r.Text())
		if line == "" || line[0] == '#' {
			continue
		}

		line = strings.SplitN(line, ",", 2)[0]

		if err := cb(line); err != nil {
			return err
		}
	}

	return nil
}

func loadFromText(u string, tr *netrie.CIDRIndex, id int16) error {
	return loadFromTextCB(u, func(s string) error {
		return tr.AddCIDR(u, id)
	})
}

// walkJSON traverses a JSON stream and calls cb for each scalar value.
func walkJSON(r io.Reader, cb func(path []string, value interface{}) error) error {
	dec := json.NewDecoder(r)
	var path []string

	// Process tokens recursively
	var traverse func() error
	traverse = func() error {
		token, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading token at path %v: %v", path, err)
		}

		switch t := token.(type) {
		case json.Delim:
			switch t {
			case '{': // Object
				for dec.More() {
					token, err := dec.Token()
					if err != nil {
						return fmt.Errorf("reading field name at path %v: %v", path, err)
					}
					field, ok := token.(string)
					if !ok {
						return fmt.Errorf("expected field name at path %v, got %v", path, token)
					}
					path = append(path, field)
					if err := traverse(); err != nil {
						return err
					}
					path = path[:len(path)-1]
				}
				// Consume closing '}'
				if _, err := dec.Token(); err != nil {
					return fmt.Errorf("reading object end at path %v: %v", path, err)
				}
			case '[': // Array
				for i := 0; dec.More(); i++ {
					path = append(path, fmt.Sprintf("%d", i))
					if err := traverse(); err != nil {
						return err
					}
					path = path[:len(path)-1]
				}
				// Consume closing ']'
				if _, err := dec.Token(); err != nil {
					return fmt.Errorf("reading array end at path %v: %v", path, err)
				}
			}
		default: // Scalar (string, number, bool, null)
			if err := cb(append([]string(nil), path...), t); err != nil {
				return err
			}
		}
		return nil
	}

	// Process the root
	return traverse()
}
