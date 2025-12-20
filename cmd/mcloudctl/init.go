package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

// InitRequest matches the server's expected request structure
type InitRequest struct {
	Name             string `json:"name"`
	AdvertiseAddress string `json:"advertise_address"`
}

// InitResponse matches the server's response structure
type InitResponse struct {
	ClusterID string `json:"cluster_id"`
	Token     string `json:"token"`
	Leader    struct {
		ID       string `json:"id"`
		Hostname string `json:"hostname"`
		IP       string `json:"ip"`
		Role     string `json:"role"`
		Status   string `json:"status"`
	} `json:"leader"`
}

// isLANInterface checks if interface name is a common LAN interface
func isLANInterface(name string) bool {
	lanPrefixes := []string{"eth", "en", "Ethernet", "Local Area Connection"}
	for _, prefix := range lanPrefixes {
		if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// isPrivateIP checks if IP is in private network range
func isPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}
	for _, cidr := range privateRanges {
		_, subnet, _ := net.ParseCIDR(cidr)
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

// getLocalIPv4 returns IPv4 address, prioritizing LAN interfaces
func getLocalIPv4() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	var lanIP, otherIP string

	for _, iface := range interfaces {
		// Skip down interfaces and loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				if ip := ipNet.IP.To4(); ip != nil {
					ipStr := ip.String()
					
					// Prioritize LAN interface with private IP
					if isLANInterface(iface.Name) && isPrivateIP(ip) {
						return ipStr, nil
					}
					
					// Store first LAN interface IP as backup
					if lanIP == "" && isLANInterface(iface.Name) {
						lanIP = ipStr
					}
					
					// Store first private IP as fallback
					if otherIP == "" && isPrivateIP(ip) {
						otherIP = ipStr
					}
				}
			}
		}
	}

	// Return best available IP
	if lanIP != "" {
		return lanIP, nil
	}
	if otherIP != "" {
		return otherIP, nil
	}
	
	return "", fmt.Errorf("no IPv4 address found")
}

// InitCommand handles the 'mcloud init' command
func InitCommand(args []string) error {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	name := fs.String("name", "", "Cluster name (required)")
	serverURL := fs.String("server", "http://127.0.0.1:9028", "mcloudd server URL")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *name == "" {
		return fmt.Errorf("--name is required")
	}

	// Auto-detect IPv4 address
	ip, err := getLocalIPv4()
	if err != nil {
		return fmt.Errorf("failed to detect IPv4 address: %w", err)
	}
	address := fmt.Sprintf("%s:8443", ip)

	fmt.Printf("Detected IPv4: %s\n", ip)
	fmt.Printf("Using address: %s\n", address)
	fmt.Println()

	// Prepare request
	req := InitRequest{
		Name:             *name,
		AdvertiseAddress: address,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send HTTP POST to mcloudd
	url := fmt.Sprintf("%s/cluster/init", *serverURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to connect to mcloudd server: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result InitResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Display result
	fmt.Println("✓ Cluster initialized successfully!")
	fmt.Println()
	fmt.Printf("Cluster ID:   %s\n", result.ClusterID)
	fmt.Printf("Cluster Name: %s\n", *name)
	fmt.Printf("Leader Node:  %s (%s)\n", result.Leader.Hostname, result.Leader.IP)
	fmt.Println()
	fmt.Println("Bootstrap Token (save this to join other nodes):")
	fmt.Printf("  %s\n", result.Token)
	fmt.Println()
	fmt.Println("To join a node, run:")
	fmt.Printf("  mcloudctl join --token %s --server %s\n", result.Token, *serverURL)

	return nil
}

// printUsage prints the CLI usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, `mcloudctl - MCloud cluster management CLI

Usage:
  mcloudctl <command> [options]

Commands:
  init    Initialize a new cluster
  join    Join an existing cluster
  help    Show this help message

Examples:
  # Initialize a new cluster
  mcloudctl init --name my-cluster

  # Join an existing cluster
  mcloudctl join --token <TOKEN>

`)
}
