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
		n1 := serviceDescs[i].ServiceName
		n2 := serviceDescs[j].ServiceName
		if n1List := strings.Split(n1, "."); len(n1List) > 4 {
			n1 = strings.Join(append(n1List[:2], n1List[3:]...), ".")
		}
		if n2List := strings.Split(n2, "."); len(n2List) > 4 {
			n2 = strings.Join(append(n2List[:2], n2List[3:]...), ".")
		}
		return n1 < n2
	})

	for _, desc := range serviceDescs {
		// Split the service name (e.g., "api.v1.user") into parts.
		keys := strings.Split(desc.ServiceName, ".")
		if len(keys) == 0 {
			continue
		}

		for _, method := range desc.Methods {
			api.tree.Routes = api.addRoute(api.tree.Routes, 0, desc.Name, method.Name, append(keys, method.MethodName))
		}
	}
}

// api.v1.merchant.wallet.add
func (api *RoutesAPI) addRoute(nodes []*RouteInfo_Node, index int, serviceName, methodName string, parts []string) []*RouteInfo_Node {
	if index >= len(parts) {
		return nodes
	}
	var target *RouteInfo_Node
	label := parts[index]
	if index == len(parts)-1 {
		label = methodName
	} else if index == len(parts)-2 {
		label = serviceName
	}

	// api.v1.auth.merchant.App.Add
	// TODO insert into admin children not auth children
	// if len(parts) > 5 && index == 3 {
	// 	// label=
	// }

	value := strings.Join(parts[:index+1], ".")
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
