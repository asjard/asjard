package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/ajerr"
	openapi_v3 "github.com/google/gnostic/openapiv3"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OpenAPI struct {
	document *openapi_v3.Document
	UnimplementedOpenAPIServer
}

func NewOpenAPI(document *openapi_v3.Document) *OpenAPI {
	return &OpenAPI{
		document: document,
	}
}

func (api *OpenAPI) Yaml(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	api.document.Info = &openapi_v3.Info{
		Title:       runtime.APP,
		Description: config.GetString(constant.ConfigDesc, ""),
		Contact: &openapi_v3.Contact{
			Name: runtime.APP,
			Url:  config.GetString(constant.ConfigWebsite, ""),
		},
		TermsOfService: config.GetString("asjard.servers.rest.openapi.termsOfService", ""),
		Version:        runtime.Version,
		License: &openapi_v3.License{
			Name: config.GetString("asjard.servers.rest.openapi.license.name", ""),
			Url:  config.GetString("asjard.servers.rest.openapi.license.url", ""),
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
			return nil, ajerr.InternalServerError
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
		if listenAddresses, ok := server.GetInstance().Endpoints[Protocol]; ok {
			if addresses := listenAddresses[constant.ServerAdvertiseAddressName]; len(addresses) > 0 {
				address = addresses[0]
			} else if addresses := listenAddresses[constant.ServerListenAddressName]; len(addresses) > 0 {
				address = addresses[0]
			}
		}
		// https://petstore.swagger.io/?url=http://%s/openapi.yml
		// https://petstore.swagger.io/?url=http://127.0.0.1:7030/openapi.yml
		// https://authress-engineering.github.io/openapi-explorer/?url=http://%s/openapi.yml
		// https://authress-engineering.github.io/openapi-explorer/?url=http://127.0.0.1:7030/openapi.yml
		redirectUrl := fmt.Sprintf(config.GetString("asjard.servers.rest.openapi.page", "https://authress-engineering.github.io/openapi-explorer/?url=http://%s/openapi.yml"), address)
		rtx.Redirect(redirectUrl, http.StatusMovedPermanently)
		return nil, nil
	}
	return &emptypb.Empty{}, nil
}

func (api OpenAPI) RestServiceDesc() *ServiceDesc {
	return &OpenAPIRestServiceDesc
}
