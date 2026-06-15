package rest

import (
	"context"
	"sort"
	"strings"
	"sync"

	"google.golang.org/protobuf/types/known/emptypb"
)

// RoutesAPI manages the discovery and hierarchical mapping of all registered routes.
type RoutesAPI struct {
	handlers []Handler // Raw list of registered service handlers.

	tree *RouteInfo // Hierarchical representation of routes (Category -> Version -> Service -> Method).
	list []string   // Flat list of full method names (e.g., "api.v1.UserService.GetUser").

	UnimplementedRoutesServer
}

var (
	routesAPI     *RoutesAPI
	routesAPIOnce sync.Once
)

// DefaultRoutes returns the singleton instance of the RoutesAPI.
func DefaultRoutes() *RoutesAPI {
	return routesAPI
}

// NewRoutesAPI initializes the singleton RoutesAPI and triggers the route generation logic.
func NewRoutesAPI(handlers []Handler) *RoutesAPI {
	routesAPIOnce.Do(func() {
		routesAPI = &RoutesAPI{
			handlers: handlers,
			tree:     &RouteInfo{},
		}
		routesAPI.genRoutes()
	})
	return routesAPI
}

// Tree returns the hierarchical route information, useful for tree-view UIs.
func (api RoutesAPI) Tree(ctx context.Context, in *emptypb.Empty) (*RouteInfo, error) {
	return api.tree, nil
}

// List returns a flat list of all available routes.
func (api RoutesAPI) List(ctx context.Context, in *emptypb.Empty) (*RouteList, error) {
	return &RouteList{Routes: api.list}, nil
}

// genRoutes processes all handlers and organizes them into a three-level tree:
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
	                    "label": "RestDefaultHandler",
	                    "value": "RestDefaultHandler"
	                }
	            ]
	        },
	        {
	            "label": "V1",
	            "value": "v1",
	            "children": [
	                {
	                    "label": "RestDefaultHandler",
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
	                    "label": "RestDefaultHandler",
	                    "value": "RestDefaultHandler"
	                }
	            ]
	        }
	    ]
	}

]
*/
// Level 1: API Type (e.g., "API")
// Level 2: Version (e.g., "V1")
// Level 3: Service & Methods (e.g., "UserService" -> "Login")
func (api *RoutesAPI) genRoutes() {
	serviceDescs := make([]*ServiceDesc, 0, len(api.handlers))
	for _, handler := range api.handlers {
		serviceDescs = append(serviceDescs, handler.RestServiceDesc())
	}

	// Sort services by name to ensure consistent output in the UI.
	sort.SliceStable(serviceDescs, func(i, j int) bool {
		return serviceDescs[i].ServiceName < serviceDescs[j].ServiceName
	})

	for _, desc := range serviceDescs {
		// Split the service name (e.g., "api.v1.user") into parts.
		keys := strings.Split(desc.ServiceName, ".")
		keyLen := len(keys)
		if keyLen == 0 {
			continue
		}

		for _, method := range desc.Methods {
			parts := make([]*nodePart, 0, len(keys)+1)
			if keyLen > 4 {
				for idx := range keyLen - 2 {
					parts = append(parts, &nodePart{
						key:   keys[idx],
						value: strings.Join(keys[:idx+1], "."),
					})
				}
				parts = append(parts, &nodePart{
					key:   keys[keyLen-1],
					value: strings.Join(keys, "."),
				})
			} else {
				for idx, item := range keys {
					parts = append(parts, &nodePart{
						key:   item,
						value: strings.Join(keys[:idx+1], "."),
					})
				}
			}

			parts = append(parts, &nodePart{key: method.MethodName, value: desc.ServiceName + "." + method.MethodName})
			api.tree.Routes = api.addRoute(api.tree.Routes, 0, desc.Name, method.Name, parts)
		}
	}
}

// api.v1.service.module.subModule

/*
	[{
		"key": "api",
		"value": "api"
		}, {
		"key": "v1",
		"value": "api.v1",
	}, {

		"key":"service",
		"value": "api.v1.service",
	}, {

		"key": "module/subModule",
	 "value": "api.v1.service.module/subModule"
	}, {

		"key": "method",
	 	"method": "method"
	}]
*/
type nodePart struct {
	key   string
	value string
}

func (api *RoutesAPI) addRoute(nodes []*RouteInfo_Node, index int, serviceName, methodName string, parts []*nodePart) []*RouteInfo_Node {
	if index >= len(parts) {
		return nodes
	}
	var target *RouteInfo_Node
	label := parts[index].key
	if index == len(parts)-1 {
		label = methodName
	} else if index == len(parts)-2 {
		label = serviceName
	}

	value := parts[index].value
	if index != len(parts)-1 {
		value += ".*"
	}

	for _, node := range nodes {
		if node.Value == value {
			target = node
			break
		}
	}
	if target == nil {
		target = &RouteInfo_Node{
			Label: label,
			Value: value,
		}
		nodes = append(nodes, target)
	}

	if index < len(parts)-1 {
		target.Children = api.addRoute(target.Children, index+1, serviceName, methodName, parts)
	}
	return nodes
}

// RestServiceDesc returns the service descriptor for the Routes introspection service.
func (RoutesAPI) RestServiceDesc() *ServiceDesc {
	return &RoutesRestServiceDesc
}
