package lists

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/vearutop/netrie"
)

// LoadFromJSONBruteForce loads CIDRs from a JSON resource, validating and adding them to the given CIDRIndex.
// The CIDRs are identified by the specified name.
// Returns an error if the resource is unreachable, JSON is invalid, or if there are issues with adding CIDRs.
func LoadFromJSONBruteForce[S int16 | int32](u string, tr *netrie.CIDRIndex[S], name string) error {
	return LoadFromJSON(
		u,
		func(path []string, value interface{}) error {
			if s, ok := value.(string); ok {
				if _, _, err := net.ParseCIDR(s); err != nil {
					return nil //nolint:nilerr // Looking for valid CIDRs in arbitrary values.
				}

				return tr.AddCIDR(s, name)
			}
			return nil
		})
}

// LoadFromJSON fetches a JSON resource from a URL and processes its scalar values using the provided callback function.
// Returns an error if the HTTP request, JSON decoding, or callback function execution fails.
func LoadFromJSON(u string, cb func(path []string, value interface{}) error) error {
	r, err := makeReader(u)
	if r != nil {
		defer r.Close()
	}

	if err != nil {
		return err
	}

	if err := walkJSON(r, cb); err != nil {
		return err
	}

	return nil
}

func makeReader(u string) (io.ReadCloser, error) {
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		f, err := os.Open(u)
		if err != nil {
			return nil, err
		}

		return f, nil
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return resp.Body, fmt.Errorf("bad HTTP status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// LoadFromTextCB loads data from a file or URL, processes each non-empty, non-comment line using the provided callback function.
// The callback receives a sanitized string extracted from each line, and errors from the callback halt further processing.
// Returns an error if any file operation, HTTP request, or callback execution fails.
func LoadFromTextCB(u string, cb func(value string) error) error {
	r, err := makeReader(u)
	if r != nil {
		defer r.Close()
	}

	if err != nil {
		return err
	}

	s := bufio.NewScanner(r)

	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || line[0] == '#' {
			continue
		}

		line = strings.SplitN(line, ",", 2)[0]

		line = strings.SplitN(line, " ", 2)[0]

		if err := cb(line); err != nil {
			return err
		}
	}

	return nil
}

// LoadFromTextGroupCIDRs loads CIDR ranges or IPs from a source, aggregates them, and adds to a netrie.Adder with a given name.
func LoadFromTextGroupCIDRs(u string, tr netrie.Adder, name string) error {
	var cidrs []string

	err := LoadFromTextCB(u, func(value string) error {
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

// LoadFromText loads CIDR ranges from a file or URL and adds them to the provided Adder with the specified name.
// It skips empty lines and comments and halts processing on the first error encountered.
// Returns an error if reading data, processing lines, or adding CIDRs fails.
func LoadFromText(u string, tr netrie.Adder, name string) error {
	return LoadFromTextCB(u, func(s string) error {
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
			return fmt.Errorf("reading token at path %v: %w", path, err)
		}

		switch t := token.(type) {
		case json.Delim:
			switch t {
			case '{': // Object
				for dec.More() {
					token, err := dec.Token()
					if err != nil {
						return fmt.Errorf("reading field name at path %v: %w", path, err)
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
					return fmt.Errorf("reading object end at path %v: %w", path, err)
				}
			case '[': // Array
				for i := 0; dec.More(); i++ {
					path = append(path, strconv.Itoa(i))
					if err := traverse(); err != nil {
						return err
					}
					path = path[:len(path)-1]
				}
				// Consume closing ']'
				if _, err := dec.Token(); err != nil {
					return fmt.Errorf("reading array end at path %v: %w", path, err)
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
