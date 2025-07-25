package rest

import (
	"context"
	"sort"
	"strings"

	"google.golang.org/protobuf/types/known/emptypb"
)

type RestDefaultHandlerAPI struct {
	handlers []Handler
	routes   *RouteInfo
	UnsafeRestDefaultHandlerServer
}

func NewRestDefaultHandlerAPI(handlers []Handler) *RestDefaultHandlerAPI {
	api := &RestDefaultHandlerAPI{
		handlers: handlers,
		routes:   &RouteInfo{},
	}
	api.genRoutes()
	return api
}

func (api RestDefaultHandlerAPI) Routes(ctx context.Context, in *emptypb.Empty) (*RouteInfo, error) {
	return api.routes, nil
}

type nodeDesc struct {
	key   string
	name  string
	value string
}

/*
[

	{
	    "label": "API",
	    "value": "api",
	    "children": [
	        {
	            "label": "V1",
	            "value": "v1",
	            "children": [
	                {
	                    "label": "rest默认方法",
	                    "value": "RestDefaultHandler"
	                }
	            ]
	        },
	        {
	            "label": "V1",
	            "value": "v1",
	            "children": [
	                {
	                    "label": "rest默认方法",
	                    "value": "RestDefaultHandler"
	                }
	            ]
	        }
	    ]
	},
	{
	    "label": "OtherAPI",
	    "value": "api",
	    "children": [
	        {
	            "label": "V1",
	            "value": "v1",
	            "children": [
	                {
	                    "label": "rest默认方法",
	                    "value": "RestDefaultHandler"
	                }
	            ]
	        }
	    ]
	}

]
*/
func (api *RestDefaultHandlerAPI) genRoutes() {
	serviceDescs := []ServiceDesc{}
	for _, handler := range api.handlers {
		serviceDescs = append(serviceDescs, *handler.RestServiceDesc())
	}
	// 排序
	sort.Slice(serviceDescs, func(i, j int) bool {
		return serviceDescs[i].ServiceName < serviceDescs[j].ServiceName
	})
	for _, desc := range serviceDescs {
		keys := strings.Split(desc.ServiceName, ".")
		keyLen := len(keys)
		if keyLen == 0 {
			continue
		}
		// api|v1|service.Handler
		// 接口类型
		tpIndex := api.routeIndex(keys[0], api.routes.Routes)
		if tpIndex < 0 {
			label := strings.ToUpper(keys[0])
			if keyLen <= 1 {
				label = desc.Name
			}
			api.routes.Routes = append(api.routes.Routes, &RouteInfo_Node{
				Label:    label,
				Value:    keys[0],
				Children: []*RouteInfo_Node{},
			})
			tpIndex = len(api.routes.Routes) - 1
		}
		if keyLen <= 1 {
			for _, method := range desc.Methods {
				api.addMethod(api.routes.Routes[tpIndex], method)
			}
			continue
		}
		// 版本号
		vIndex := api.routeIndex(keys[1], api.routes.Routes[tpIndex].Children)
		if vIndex < 0 {
			label := strings.ToUpper(keys[1])
			if len(keys) <= 2 {
				label = desc.Name
			}
			api.addRoute(api.routes.Routes[tpIndex], label, keys[1])
			vIndex = len(api.routes.Routes[tpIndex].Children) - 1
		}
		if len(keys) <= 2 {
			for _, method := range desc.Methods {
				api.addMethod(api.routes.Routes[tpIndex].Children[vIndex], method)
			}
			continue
		}
		// 服务
		sName := strings.Join(keys[2:], ".")
		sIndex := api.routeIndex(sName, api.routes.Routes[tpIndex].Children[vIndex].Children)
		if sIndex < 0 {
			api.addRoute(api.routes.Routes[tpIndex].Children[vIndex], desc.Name, sName)
			sIndex = len(api.routes.Routes[tpIndex].Children[vIndex].Children) - 1
		}

		// 方法
		for _, method := range desc.Methods {
			api.addMethod(api.routes.Routes[tpIndex].Children[vIndex].Children[sIndex], method)
		}

	}
}

func (api *RestDefaultHandlerAPI) routeIndex(value string, nodes []*RouteInfo_Node) int {
	for index, route := range nodes {
		if route.Value == value {
			return index
		}
	}
	return -1
}

func (api *RestDefaultHandlerAPI) addMethod(node *RouteInfo_Node, method MethodDesc) {
	label := method.Name
	if label == "" {
		label = method.MethodName
	}
	node.Children = append(node.Children, &RouteInfo_Node{
		Label: label,
		Value: method.MethodName,
	})
}

func (api *RestDefaultHandlerAPI) addRoute(node *RouteInfo_Node, label, value string) {
	if label == "" {
		keys := strings.Split(value, ".")
		if len(keys) > 0 {
			label = keys[len(keys)-1]
		}
	}
	node.Children = append(node.Children, &RouteInfo_Node{
		Label:    label,
		Value:    value,
		Children: []*RouteInfo_Node{},
	})
}

func (RestDefaultHandlerAPI) RestServiceDesc() *ServiceDesc {
	return &RestDefaultHandlerRestServiceDesc
}
