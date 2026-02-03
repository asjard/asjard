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
	serviceDescs := []ServiceDesc{}
	for _, handler := range api.handlers {
		serviceDescs = append(serviceDescs, *handler.RestServiceDesc())
	}

	// Sort services by name to ensure consistent output in the UI.
	sort.Slice(serviceDescs, func(i, j int) bool {
		return serviceDescs[i].ServiceName < serviceDescs[j].ServiceName
	})

	for _, desc := range serviceDescs {
		// Split the service name (e.g., "api.v1.user") into parts.
		keys := strings.Split(desc.ServiceName, ".")
		keyLen := len(keys)
		if keyLen == 0 {
			continue
		}

		// 1. Handle API Type Level (e.g., "api")
		tpIndex := api.routeIndex(keys[0], api.tree.Routes)
		if tpIndex < 0 {
			label := strings.ToUpper(keys[0])
			if keyLen <= 1 {
				label = desc.Name
			}
			api.tree.Routes = append(api.tree.Routes, &RouteInfo_Node{
				Label:    label,
				Value:    keys[0],
				Children: []*RouteInfo_Node{},
			})
			tpIndex = len(api.tree.Routes) - 1
		}
		if keyLen <= 1 {
			for _, method := range desc.Methods {
				api.addMethod(api.tree.Routes[tpIndex], method)
			}
			continue
		}

		// 2. Handle Version Level (e.g., "v1")
		vIndex := api.routeIndex(keys[1], api.tree.Routes[tpIndex].Children)
		if vIndex < 0 {
			label := strings.ToUpper(keys[1])
			if len(keys) <= 2 {
				label = desc.Name
			}
			api.addRoute(api.tree.Routes[tpIndex], label, keys[1])
			vIndex = len(api.tree.Routes[tpIndex].Children) - 1
		}
		if len(keys) <= 2 {
			for _, method := range desc.Methods {
				api.addMethod(api.tree.Routes[tpIndex].Children[vIndex], method)
			}
			continue
		}

		// 3. Handle Service Name Level
		sName := strings.ReplaceAll(strings.Join(keys[2:], "."), ".", "_")
		sIndex := api.routeIndex(sName, api.tree.Routes[tpIndex].Children[vIndex].Children)
		if sIndex < 0 {
			api.addRoute(api.tree.Routes[tpIndex].Children[vIndex], desc.Name, sName)
			sIndex = len(api.tree.Routes[tpIndex].Children[vIndex].Children) - 1
		}

		// 4. Add Methods as leaf nodes
		for _, method := range desc.Methods {
			api.addMethod(api.tree.Routes[tpIndex].Children[vIndex].Children[sIndex], method)
			api.list = append(api.list, desc.ServiceName+"."+method.MethodName)
		}
	}
}

// routeIndex is a helper to find a node in a slice by its 'Value' property.
func (api *RoutesAPI) routeIndex(value string, nodes []*RouteInfo_Node) int {
	for index, route := range nodes {
		if route.Value == value {
			return index
		}
	}
	return -1
}

// addMethod adds a method node to a parent route node.
func (api *RoutesAPI) addMethod(node *RouteInfo_Node, method MethodDesc) {
	label := method.Name
	if label == "" {
		label = method.MethodName
	}
	node.Children = append(node.Children, &RouteInfo_Node{
		Label: label,
		Value: method.MethodName,
	})
}

// addRoute adds a category or service node to a parent route node.
func (api *RoutesAPI) addRoute(node *RouteInfo_Node, label, value string) {
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

// RestServiceDesc returns the service descriptor for the Routes introspection service.
func (RoutesAPI) RestServiceDesc() *ServiceDesc {
	return &RoutesRestServiceDesc
}
