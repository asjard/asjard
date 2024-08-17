package utils

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// GetListenAddress 获取本地监听地址,
// 如果host为0.0.0.0或::则返回本地IPv4或者IPv6地址
func GetListenAddress(hostPort string) (string, error) {
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
	return net.JoinHostPort(addr, port), nil
}

// LocalIPv4 本地IPV4地址
func LocalIPv4() string {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addresses {
		// Parse IP
		var ip net.IP
		if ip, _, err = net.ParseCIDR(address.String()); err != nil {
			return ""
		}
		// Check if valid global unicast IPv4 address
		if ip != nil && (ip.To4() != nil) && ip.IsGlobalUnicast() {
			return ip.String()
		}
	}
	return ""
}

// LocalIPv6 本地IPv6地址
func LocalIPv6() string {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addresses {
		// Parse IP
		var ip net.IP
		if ip, _, err = net.ParseCIDR(address.String()); err != nil {
			return ""
		}
		// Check if valid IPv6 address
		if ip != nil && (ip.To16() != nil) && IsIPv6(ip) && ip.IsGlobalUnicast() {
			return ip.String()
		}
	}
	return ""
}

// IsIPv6 是否IPv6地址
func IsIPv6(ip net.IP) bool {
	if ip != nil && strings.Contains(ip.String(), ":") {
		return true
	}
	return false
}

// ParseAddress 解析监听地址为host和port
func ParseAddress(address string) (string, int, error) {
	var portInt int
	host, port, err := net.SplitHostPort(address)
	if err == nil {
		portInt, err = strconv.Atoi(port)
	}
	return host, portInt, err
}
