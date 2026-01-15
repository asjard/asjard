package server

import (
	"errors"
	"sync"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/utils"
)

// Service represents the full profile of a running service instance.
// It combines organizational metadata (APP) with network access points (Endpoints).
type Service struct {
	// runtime.APP provides identity: App Name, Environment, Region, and Instance ID.
	runtime.APP

	// Endpoints maps protocol names (e.g., "grpc", "http") to their network addresses.
	// Key: protocol name (string)
	// Value: Endpoint struct containing listen and advertise addresses.
	Endpoints map[string]*Endpoint

	// em protects concurrent access to the Endpoints map.
	em sync.RWMutex
}

// service is the global singleton representing the current running process.
var service *Service
var sonce sync.Once

// GetService returns the singleton service instance, initializing it if necessary.
// It ensures that the application has a consistent identity across all modules.
func GetService() *Service {
	sonce.Do(func() {
		service = NewService()
	})
	return service
}

// NewService creates a new Service container and populates it with
// basic application metadata from the runtime package.
func NewService() *Service {
	return &Service{
		APP:       runtime.GetAPP(),
		Endpoints: make(map[string]*Endpoint),
	}
}

// AddEndpoint registers a network address for a specific protocol.
// It automatically resolves local IP addresses if a port-only string (like ":8080") is provided.
func (s *Service) AddEndpoint(protocol string, address AddressConfig) error {
	if protocol == "" {
		return errors.New("endpoint protocol is must")
	}

	// Ensure the protocol entry exists in the map.
	s.em.RLock()
	if _, ok := s.Endpoints[protocol]; !ok {
		s.em.RUnlock()
		s.em.Lock()
		if _, ok := s.Endpoints[protocol]; !ok {
			s.Endpoints[protocol] = &Endpoint{}
		}
		s.em.Unlock()
	} else {
		s.em.RUnlock()
	}

	s.em.Lock()
	defer s.em.Unlock()

	// Process the Listen address.
	if address.Listen != "" {
		// utils.GetListenAddress helps resolve real IPs from shorthand notation.
		listenAddress, err := utils.GetListenAddress(address.Listen)
		if err != nil {
			return err
		}
		logger.Debug("service listen address", "protocol", protocol, "listen", listenAddress)
		s.Endpoints[protocol].Listen = append(s.Endpoints[protocol].Listen, listenAddress)
	}

	// Process the Advertise address (the one shared with service discovery).
	if address.Advertise != "" {
		s.Endpoints[protocol].Advertise = append(s.Endpoints[protocol].Advertise, address.Advertise)
	}
	return nil
}

// GetListenAddresses retrieves all local binding addresses for a specific protocol.
func (s *Service) GetListenAddresses(protocol string) []string {
	s.em.RLock()
	defer s.em.RUnlock()
	endpoint, ok := s.Endpoints[protocol]
	if !ok {
		return []string{}
	}
	return endpoint.Listen
}

// GetAdvertiseAddresses retrieves the addresses intended for public/external discovery.
func (s *Service) GetAdvertiseAddresses(protocol string) []string {
	s.em.RLock()
	defer s.em.RUnlock()
	endpoint, ok := s.Endpoints[protocol]
	if !ok {
		return []string{}
	}
	return endpoint.Advertise
}

// GetEndpoint returns the full Endpoint object for a protocol.
func (s *Service) GetEndpoint(protocol string) (*Endpoint, bool) {
	s.em.RLock()
	defer s.em.RUnlock()
	endpoint, ok := s.Endpoints[protocol]
	return endpoint, ok
}
