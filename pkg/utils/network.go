package utils

import (
	"fmt"
	"net"
)

// IsLANInterface checks if the given network interface name is a common LAN interface.
// It checks against common prefixes used by LAN adapters on different operating systems:
//   - "eth" (Linux Ethernet)
//   - "en" (macOS/BSD Ethernet)
//   - "Ethernet" (Windows)
//   - "Local Area Connection" (Windows legacy)
//
// Parameters:
//   name - The network interface name to check
//
// Returns:
//   true if the interface name matches a LAN interface pattern, false otherwise
func IsLANInterface(name string) bool {
	lanPrefixes := []string{"eth", "en", "Ethernet", "Local Area Connection"}
	// Check if the interface name starts with any of the LAN prefixes
	for _, prefix := range lanPrefixes {
		if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// IsPrivateIP checks if the given IP address is in a private network range.
// Private IP ranges are defined by RFC 1918:
//   - 10.0.0.0/8        (10.0.0.0 - 10.255.255.255)
//   - 172.16.0.0/12     (172.16.0.0 - 172.31.255.255)
//   - 192.168.0.0/16    (192.168.0.0 - 192.168.255.255)
//
// Parameters:
//   ip - The IP address to check
//
// Returns:
//   true if the IP is in a private range, false otherwise
func IsPrivateIP(ip net.IP) bool {
	privateRanges := []string{
		"10.0.0.0/8",      // Class A private network
		"172.16.0.0/12",   // Class B private network
		"192.168.0.0/16",  // Class C private network
	}
	// Check if the IP falls within any of the private ranges
	for _, cidr := range privateRanges {
		_, subnet, _ := net.ParseCIDR(cidr)
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

// GetLocalIPv4 returns the local IPv4 address with a priority system for selecting the best address.
// 
// Priority order:
//   1. LAN interface (eth*, en*, Ethernet*) with private IP (highest priority)
//   2. Any LAN interface IP (if private IP not found)
//   3. Any private IP from other interfaces (fallback)
//
// This function is useful for automatically detecting the machine's IP address for cluster communication,
// ensuring that it prefers actual network adapters over virtual/bridge interfaces.
//
// Returns:
//   - The selected IPv4 address as a string
//   - An error if no suitable IPv4 address is found
func GetLocalIPv4() (string, error) {
	// Get all network interfaces on the system
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	var lanIP, otherIP string

	// Iterate through all network interfaces
	for _, iface := range interfaces {
		// Skip interfaces that are not up or are loopback devices
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Get all addresses associated with this interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Check each address on the interface
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok {
				// Convert to IPv4 (returns nil if IPv6)
				if ip := ipNet.IP.To4(); ip != nil {
					ipStr := ip.String()
					
					// Best case: LAN interface with private IP - return immediately
					if IsLANInterface(iface.Name) && IsPrivateIP(ip) {
						return ipStr, nil
					}
					
					// Second priority: Store first LAN interface IP as backup
					if lanIP == "" && IsLANInterface(iface.Name) {
						lanIP = ipStr
					}
					
					// Third priority: Store first private IP as fallback
					if otherIP == "" && IsPrivateIP(ip) {
						otherIP = ipStr
					}
				}
			}
		}
	}

	// Return best available IP based on priority
	if lanIP != "" {
		return lanIP, nil
	}
	if otherIP != "" {
		return otherIP, nil
	}
	
	return "", fmt.Errorf("no IPv4 address found")
}

// GetAllIPs returns a list of all IPv4 addresses from active network interfaces on the system.
// This function excludes loopback interfaces and only returns IPv4 addresses (not IPv6).
//
// Useful for discovering all available IP addresses on the machine, such as for:
//   - Network diagnostics
//   - Displaying available interfaces to users
//   - Cluster node discovery
//
// Returns:
//   A slice of net.IP containing all IPv4 addresses found, or an empty slice if none or error occurs
func GetAllIPs() []net.IP {
	var ips []net.IP

	// Get all network interfaces on the system
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range ifaces {
		// Skip interfaces that are down (not active)
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		// Skip loopback interfaces (127.0.0.1)
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Get all addresses associated with this interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP

			// Extract IP address from the address type
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil {
				continue
			}

			// Convert to IPv4 only (returns nil if it's IPv6)
			ip = ip.To4()
			if ip == nil {
				continue
			}

			// Add the IPv4 address to the result list
			ips = append(ips, ip)
		}
	}

	return ips
}
