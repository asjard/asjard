package rest

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/status"
	scalargo "github.com/bdpiprava/scalar-go"
	"github.com/ghodss/yaml"
	openapi_v3 "github.com/google/gnostic/openapiv3"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/types/known/emptypb"
)

// OpenAPI handles the generation and serving of API specifications.
type OpenAPI struct {
	document  *openapi_v3.Document
	mergeOnce sync.Once // Ensures metadata merging happens only once.
	conf      OpenapiConfig
	service   *server.Service
	UnimplementedOpenAPIServer
}

// NewOpenAPI initializes a new OpenAPI provider with the base document and configuration.
func NewOpenAPI(conf OpenapiConfig, document *openapi_v3.Document) *OpenAPI {
	return &OpenAPI{
		document: document,
		conf:     conf,
		service:  server.GetService(),
	}
}

// Yaml returns the OpenAPI specification in YAML format.
func (api *OpenAPI) Yaml(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	api.mergeOnce.Do(func() { api.merge() })
	rtx, ok := ctx.(*Context)
	if ok {
		// Generate YAML with a framework-specific header comment.
		b, err := api.document.YAMLValue(fmt.Sprintf("Generated with %s(%s) \n %s",
			constant.Framework, constant.FrameworkVersion, constant.FrameworkGithub))
		if err != nil {
			logger.Error("openapi yaml value fail", "err", err)
			return nil, status.InternalServerError()
		}
		rtx.RequestCtx.Write(b)
		return nil, nil
	}
	return &emptypb.Empty{}, nil
}

// Json returns the OpenAPI specification in JSON format.
func (api *OpenAPI) Json(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	api.mergeOnce.Do(func() { api.merge() })
	rtx, ok := ctx.(*Context)
	if ok {
		b, err := api.document.YAMLValue(fmt.Sprintf("Generated with %s(%s) \n %s",
			constant.Framework, constant.FrameworkVersion, constant.FrameworkGithub))
		if err != nil {
			logger.Error("openapi yaml value fail", "err", err)
			return nil, status.InternalServerError()
		}
		// Convert the YAML representation to JSON.
		jb, err := yaml.YAMLToJSON(b)
		if err != nil {
			logger.Error("convert openapi yaml to json fail", "err", err)
			return nil, status.InternalServerError()
		}
		rtx.RequestCtx.Write(jb)
		return nil, nil
	}
	return &emptypb.Empty{}, nil
}

// Page redirects users to the external Swagger UI page defined in the config.
func (api *OpenAPI) Page(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	rtx, ok := ctx.(*Context)
	if ok {
		rtx.Redirect(fmt.Sprintf(api.conf.Page, api.listenAddress()), http.StatusMovedPermanently)
		return nil, nil
	}
	return &emptypb.Empty{}, nil
}

// ScalarPage renders an embedded, high-performance Scalar API Reference page.
func (api *OpenAPI) ScalarPage(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	rtx, ok := ctx.(*Context)
	if ok {
		content, err := scalargo.NewV2(api.scalarOptions()...)
		if err != nil {
			logger.Error("new scalargo fail", "err", err)
			return &emptypb.Empty{}, status.InternalServerError()
		}
		rtx.Response.Header.Set(fasthttp.HeaderContentType, "text/html")
		rtx.WriteString(content)
		return nil, nil
	}
	return &emptypb.Empty{}, nil
}

// scalarOptions builds the UI configuration for the Scalar reference tool.
func (api *OpenAPI) scalarOptions() []scalargo.Option {
	options := []scalargo.Option{
		scalargo.WithSpecURL(api.listenAddress() + "/openapi.yaml"),
	}
	// Apply UI themes, CDN settings, and visibility toggles from the config.
	if api.conf.Scalar.Theme != "" {
		options = append(options, scalargo.WithTheme(scalargo.Theme(api.conf.Scalar.Theme)))
	}

	if api.conf.Scalar.CDN != "" {
		options = append(options, scalargo.WithCDN(api.conf.Scalar.CDN))
	}

	if api.conf.Scalar.SidebarVisibility {
		options = append(options, scalargo.WithSidebarVisibility(api.conf.Scalar.SidebarVisibility))
	}

	if api.conf.Scalar.HideModels {
		options = append(options, scalargo.WithHideModels())
	}

	if api.conf.Scalar.HideDownloadButton {
		options = append(options, scalargo.WithHideDownloadButton())
	}

	if api.conf.Scalar.DarkMode {
		options = append(options, scalargo.WithDarkMode())
	}

	if len(api.conf.Scalar.HideClients) != 0 {
		allClients := false
		for _, client := range api.conf.Scalar.HideClients {
			if client == "*" {
				allClients = true
				break
			}
		}
		if allClients {
			options = append(options, scalargo.WithHideAllClients())
		} else {
			options = append(options, scalargo.WithHiddenClients(api.conf.Scalar.HideClients...))
		}
	}

	if api.conf.Scalar.Authentication != "" {
		options = append(options, scalargo.WithAuthentication(api.conf.Scalar.Authentication))
	}

	options = append(options, scalargo.WithBaseServerURL(api.listenAddress()))
	return options
}

// listenAddress determines the base URL for the API, prioritizing the configured Endpoint.
func (api *OpenAPI) listenAddress() string {
	if api.conf.Endpoint != "" {
		return api.conf.Endpoint
	}
	address := ""
	if addresses := api.service.GetAdvertiseAddresses(Protocol); len(addresses) > 0 {
		address = addresses[0]
	} else if addresses := api.service.GetListenAddresses(Protocol); len(addresses) > 0 {
		address = addresses[0]
	}
	if address != "" && !strings.HasPrefix(address, "http") {
		address = "http://" + address
	}
	return address
}

// merge synchronizes the OpenAPI document's Info block with the current service metadata.
func (api *OpenAPI) merge() {
	api.document.Info = &openapi_v3.Info{
		Title:       api.service.App,
		Description: api.service.Desc,
		Contact: &openapi_v3.Contact{
			Name: api.service.App,
			Url:  api.service.Website,
		},
		TermsOfService: api.conf.TermsOfService,
		Version:        api.service.Instance.Version,
		License: &openapi_v3.License{
			Name: api.conf.License.Name,
			Url:  api.conf.License.Url,
		},
	}

	// Deduplicate schemas to prevent conflicts in the final document.
	props := make([]*openapi_v3.NamedSchemaOrReference, 0, len(api.document.Components.Schemas.AdditionalProperties))
	propMap := make(map[string]struct{})
	for _, prop := range api.document.Components.Schemas.AdditionalProperties {
		if _, ok := propMap[prop.Name]; !ok {
			props = append(props, prop)
			propMap[prop.Name] = struct{}{}
		}
	}
	api.document.Components.Schemas.AdditionalProperties = props

	// Set the server URL so interactive "Try it out" features work correctly.
	api.document.Servers = []*openapi_v3.Server{{
		Url: api.listenAddress(),
	}}
}

// RestServiceDesc returns the service description for REST registration.
func (api *OpenAPI) RestServiceDesc() *ServiceDesc {
	return &OpenAPIRestServiceDesc
}
