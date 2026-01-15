package utils

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// GetListenAddress resolves a local listening address.
// If the provided host is an unspecified address (0.0.0.0 for IPv4 or :: for IPv6),
// it retrieves the first available global unicast IP address from the local machine's interfaces.
func GetListenAddress(hostPort string) (string, error) {
	// Splitting the input into host and port components.
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return "", err
	}
	var addr string
	ip := net.ParseIP(host)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address %s", host)
	}

	addr = host
	// If the IP is "unspecified" (0.0.0.0 or ::), we need to find a real local IP
	// that other services can use to reach this node.
	if ip.IsUnspecified() {
		if IsIPv6(ip) {
			addr = LocalIPv6()
		} else {
			addr = LocalIPv4()
		}
		if addr == "" {
			return addr, errors.New("failed to get local IP address")
		}
	}
	// Reconstruct the address with the resolved IP and the original port.
	return net.JoinHostPort(addr, port), nil
}

// LocalIPv4 scans local network interfaces and returns the first valid,
// global unicast IPv4 address found.
func LocalIPv4() string {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addresses {
		var ip net.IP
		// Extract IP from CIDR notation (e.g., "192.168.1.5/24").
		if ip, _, err = net.ParseCIDR(address.String()); err != nil {
			return ""
		}
		// We filter for Global Unicast to avoid loopback (127.0.0.1)
		// or link-local addresses that aren't routable across the network.
		if ip != nil && (ip.To4() != nil) && ip.IsGlobalUnicast() {
			return ip.String()
		}
	}
	return ""
}

// LocalIPv6 scans local network interfaces and returns the first valid,
// global unicast IPv6 address found.
func LocalIPv6() string {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addresses {
		var ip net.IP
		if ip, _, err = net.ParseCIDR(address.String()); err != nil {
			return ""
		}
		// Ensures the IP is a 16-byte IPv6 representation and is globally routable.
		if ip != nil && (ip.To16() != nil) && IsIPv6(ip) && ip.IsGlobalUnicast() {
			return ip.String()
		}
	}
	return ""
}

// IsIPv6 is a helper to determine if a net.IP is version 6 based on string representation.
func IsIPv6(ip net.IP) bool {
	// IPv6 addresses contain colons, whereas IPv4 addresses contain dots.
	if ip != nil && strings.Contains(ip.String(), ":") {
		return true
	}
	return false
}

// ParseAddress breaks down a "host:port" string into its constituent parts,
// converting the port into an integer for programmatic use.
func ParseAddress(address string) (string, int, error) {
	var portInt int
	host, port, err := net.SplitHostPort(address)
	if err == nil {
		portInt, err = strconv.Atoi(port)
	}
	return host, portInt, err
}
