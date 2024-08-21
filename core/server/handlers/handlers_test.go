package handlers

import (
	"testing"

	"github.com/asjard/asjard/core/constant"
)

type testDefaultHandler struct{}

func TestDefaultHandlers(t *testing.T) {
	t.Run("AddHandler", func(t *testing.T) {
		datas := []struct {
			name      string
			handler   any
			protocols []string
			exist     bool
		}{
			{name: "test_default_handler", handler: &testDefaultHandler{}, protocols: []string{"rest"}, exist: true},
			{name: "test_default_handler_1", handler: &testDefaultHandler{}, protocols: []string{"rest", "rest1"}, exist: true},
			{name: "test_default_handler_2", handler: &testDefaultHandler{}, protocols: []string{"rest", "grpc"}, exist: true},
			{name: "test_default_handler_3", handler: &testDefaultHandler{}, protocols: []string{}, exist: true},
		}
		for _, data := range datas {
			AddServerDefaultHandler(data.name, data.handler, data.protocols...)
			if len(data.protocols) == 0 {
				data.protocols = []string{constant.AllProtocol}
			}
			for _, protocol := range data.protocols {
				if _, ok := serverDefaultHandlers[protocol]; ok != data.exist {
					t.Errorf("handler %s actually %v, current %v", data.name, data.exist, ok)
					t.FailNow()
				}
			}
		}
	})
	t.Run("GetHandlers", func(t *testing.T) {

	})
}
