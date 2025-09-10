package lists

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/vearutop/netrie"
)

func loadFromJSONBruteForce(u string, tr *netrie.CIDRIndex, name string) error {
	return loadFromJSON(
		u,
		func(path []string, value interface{}) error {
			if s, ok := value.(string); ok {
				_, _, err := net.ParseCIDR(s)
				if err != nil {
					return nil
				}

				return tr.AddCIDR(s, name)
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
	var r io.Reader
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		f, err := os.Open(u)
		if err != nil {
			return err
		}
		defer f.Close()

		r = f
	} else {
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

		r = resp.Body
	}

	s := bufio.NewScanner(r)

	for s.Scan() {
		line := strings.TrimSpace(s.Text())
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

func loadFromTextGroupCIDRs(u string, tr *netrie.CIDRIndex, name string) error {
	var cidrs []string

	err := loadFromTextCB(u, func(value string) error {
		cidrs = append(cidrs, value)
		return nil
	})
	if err != nil {
		return err
	}

	nets, err := ClusterCIDRs(cidrs)
	if err != nil {
		return err
	}

	for _, n := range nets {
		tr.AddNet(n, name)
	}

	return nil
}

func loadFromText(u string, tr *netrie.CIDRIndex, name string) error {
	return loadFromTextCB(u, func(s string) error {
		return tr.AddCIDR(s, name)
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
