package rest

import (
	"fmt"
	"net/http"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/server"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

const (
	// Protocol 协议名称
	Protocol = "rest"
)

const (
	defaultReadBufferSize  = 4096
	defaultWriteBufferSize = 4096
)

// Handler .
type Handler interface {
	Routers() []*Router
	Groups() []*Group
}

// RestServer .
type RestServer struct {
	addresses map[string]string
	router    *routing.Router
	server    fasthttp.Server
}

var _ server.Server = &RestServer{}

func init() {
	server.AddServer(New)
}

// New .
func New() (server.Server, error) {
	addressesMap := make(map[string]string)
	if err := config.GetWithUnmarshal("servers.http.addresses", &addressesMap); err != nil {
		return nil, err
	}
	return &RestServer{
		addresses: addressesMap,
		router:    routing.New(),
		server: fasthttp.Server{
			Concurrency:    fasthttp.DefaultConcurrency,
			ReadBufferSize: defaultReadBufferSize,
		},
	}, nil
}

// AddHandler .
func (s *RestServer) AddHandler(handler any) error {
	h, ok := handler.(Handler)
	if !ok {
		return fmt.Errorf("invlaid handler, must implement *rest.Handler")
	}
	if err := s.addRouter(h.Routers()); err != nil {
		return err
	}
	return nil
}

// Start .
func (s *RestServer) Start() error {
	return nil
	// return s.server.ListenAndServe("")
}

// Stop .
func (s *RestServer) Stop() {
}

// Protocol .
func (s *RestServer) Protocol() string {
	return Protocol
}

// ListenAddresses 监听地址列表
func (s *RestServer) ListenAddresses() []*server.EndpointAddress {
	var addresses []*server.EndpointAddress
	for name, address := range s.addresses {
		addresses = append(addresses, &server.EndpointAddress{
			Name:    name,
			Address: address,
		})
	}
	return addresses
}

func (s *RestServer) addRouter(routers []*Router) error {
	for _, router := range routers {
		s.router.To(router.Method, router.Path, func(ctx *routing.Context) error {
			switch router.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch:
			default:
			}
			return nil
		})
	}
	return nil
}
