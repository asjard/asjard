package client

import (
	"sync"

	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/metadata"
)

const (
	HeaderLoadBalancer = "x-request-balancer"
)

// Picker 所有的负载句衡器都要实现的接口
type Picker interface {
	Pick(info balancer.PickInfo) (*PickResult, error)
	Name() string
}

// PickResult 负载均衡结果
type PickResult struct {
	SubConn       *SubConn
	RequestRegion string
	RequestAz     string
}

// WrapPicker 在本身的选择器上再添加一层，统一返回pickResult
type WrapPicker struct {
	app runtime.APP

	// picker参数由mu锁保护
	mu     sync.Mutex
	picker Picker
}

type PickerBase struct {
	app runtime.APP
}

func NewPickerBase() *PickerBase {
	return &PickerBase{app: runtime.GetAPP()}
}

// CanReachable 是否可达
func (p PickerBase) CanReachable(sc *SubConn) bool {
	instance, ok := sc.Address.Attributes.Value(AddressAttrKey{}).(*registry.Instance)
	// 同region listen
	// 或者不同region，advertise
	if ok {
		if sc.Address.Attributes.Value(ListenAddressKey{}) != nil && instance.Service.Region == p.app.Region {
			return true
		}
		if sc.Address.Attributes.Value(AdvertiseAddressKey{}) != nil && instance.Service.Region != p.app.Region {
			return true
		}
	}
	return false
}

// Shareable 是否被共享
func (p PickerBase) Shareable(sc *SubConn) bool {
	instance, ok := sc.Address.Attributes.Value(AddressAttrKey{}).(*registry.Instance)
	if ok {
		return instance.Service.Instance.Shareable
	}
	return false
}

func NewPicker(newPicker NewBalancerPicker, scs map[balancer.SubConn]base.SubConnInfo) balancer.Picker {
	return &WrapPicker{
		picker: newPicker(scs),
		app:    runtime.GetAPP(),
	}
}

func (p *WrapPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.mu.Lock()
	result, err := p.picker.Pick(info)
	p.mu.Unlock()
	if err != nil {
		return balancer.PickResult{}, err
	}
	destServicename := ""
	instance, ok := result.SubConn.Address.Attributes.Value(AddressAttrKey{}).(*registry.Instance)
	if ok {
		destServicename = instance.Service.Instance.Name
	}
	requestRegion := result.RequestRegion
	if requestRegion == "" {
		requestRegion = p.app.Region
	}
	requestAz := result.RequestAz
	if requestAz == "" {
		requestAz = p.app.AZ
	}
	return balancer.PickResult{
		SubConn: result.SubConn.Conn,
		Done:    func(info balancer.DoneInfo) {},
		Metadata: metadata.New(map[string]string{
			HeaderRequestApp:    p.app.App,
			HeaderRequestSource: p.app.Instance.Name,
			HeaderRequestDest:   destServicename,
			HeaderRequestRegion: requestRegion,
			HeaderRequestAz:     requestAz,
			HeaderLoadBalancer:  p.picker.Name(),
		}),
	}, nil
}
