package server

// Endpoint represents the network access points for a specific communication protocol.
// A single service may have multiple endpoints (e.g., one for internal gRPC and one for public REST).
type Endpoint struct {
	// Listen is a list of local network addresses the server is currently bound to.
	// Format is typically "IP:Port" or ":Port" (e.g., "0.0.0.0:8080").
	// Multiple addresses allow the server to listen on different interfaces (IPv4/IPv6/Localhost).
	Listen []string

	// Advertise is a list of addresses published to the service registry (e.g., ETCD, Consul).
	// These are the addresses that clients will use to reach this service.
	// In environments like Docker or Kubernetes, this might be a Load Balancer IP
	// or a NodePort, which differs from the internal 'Listen' address.
	Advertise []string
}
