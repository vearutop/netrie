package lists

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/vearutop/netrie"
)

// LoadCloud fills CIDRIndex with networks from
// https://github.com/disposable/cloud-ip-ranges/tree/master/txt.
func LoadCloud(tr *netrie.CIDRIndex) error {
	apiURL := "https://api.github.com/repos/disposable/cloud-ip-ranges/contents/txt"

	type FileEntry struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		DownloadURL string `json:"download_url"`
	}

	// Request directory listing
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("listing directory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed: %s", resp.Status)
	}

	var entries []FileEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	for _, entry := range entries {
		if entry.Type != "file" {
			continue
		}

		name := strings.TrimSuffix(entry.Name, ".txt")

		if name == "ahrefs" {
			if err := loadFromTextGroupIPs(entry.DownloadURL, tr, name); err != nil {
				return err
			}

			continue
		}

		if name == "apple-icloud" {
			if err := loadFromTextGroupCIDRs(entry.DownloadURL, tr, name); err != nil {
				return err
			}

			continue
		}

		if err := loadFromText(entry.DownloadURL, tr, name); err != nil {
			return err
		}
	}

	return nil
}
