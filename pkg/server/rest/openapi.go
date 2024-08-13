package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/status"
	openapi_v3 "github.com/google/gnostic/openapiv3"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OpenAPI struct {
	document *openapi_v3.Document
	conf     OpenapiConfig
	service  *server.Service
	UnimplementedOpenAPIServer
}

func NewOpenAPI(conf OpenapiConfig, document *openapi_v3.Document) *OpenAPI {
	return &OpenAPI{
		document: document,
		conf:     conf,
		service:  server.GetService(),
	}
}

func (api *OpenAPI) Yaml(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
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
	props := make([]*openapi_v3.NamedSchemaOrReference, 0, len(api.document.Components.Schemas.AdditionalProperties))
	propMap := make(map[string]struct{})
	for _, prop := range api.document.Components.Schemas.AdditionalProperties {
		if _, ok := propMap[prop.Name]; !ok {
			props = append(props, prop)
		}
		propMap[prop.Name] = struct{}{}
	}
	api.document.Components.Schemas.AdditionalProperties = props
	rtx, ok := ctx.(*Context)
	if ok {
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

func (api *OpenAPI) Page(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	rtx, ok := ctx.(*Context)
	if ok {
		var address string
		if addresses := api.service.GetAdvertiseAddresses(Protocol); len(addresses) > 0 {
			address = addresses[0]
		} else if addresses := api.service.GetListenAddresses(Protocol); len(addresses) > 0 {
			address = addresses[0]
		}
		rtx.Redirect(fmt.Sprintf(api.conf.Page, address), http.StatusMovedPermanently)
		return nil, nil
	}
	return &emptypb.Empty{}, nil
}

func (api OpenAPI) RestServiceDesc() *ServiceDesc {
	return &OpenAPIRestServiceDesc
}
