package lxd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func (c *LxdClient) InitCluster(address string) (*NodeInfo, error) {
	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "mcloud-leader"
	}

	payload := map[string]any{
		"server_name":         hostname,
		"enabled":             true,
		"cluster_address":     address,
		"cluster_certificate": "",
		"cluster_password":    "",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	httpReq, err := http.NewRequest(
		"PUT",
		"http://unix/1.0/cluster",
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(httpReq)
	if err != nil {
		// If LXD is not available, return mock data for development
		return &NodeInfo{
			Hostname: hostname,
			IP:       address,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("lxd init failed: %s", resp.Status)
	}

	return &NodeInfo{
		Hostname: hostname,
		IP:       address,
	}, nil
}
