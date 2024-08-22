package utils

import (
	"net/http"
	"strings"

	"github.com/asjard/asjard/pkg/protobuf/httppb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

const (
	pathDelimiter = "/"
)

// HttpOption http请求参数
type HttpOption struct {
	Path   string
	Method string
	Body   string

	Api        string
	Version    string
	Group      string
	WriterName string
}

// Path 请求路径
func (h HttpOption) GetPath() string {
	return strings.TrimSuffix(strings.Join([]string{
		"",
		strings.Trim(h.Api, pathDelimiter),
		strings.Trim(h.Version, pathDelimiter),
		strings.Trim(h.Group, pathDelimiter),
		strings.Trim(h.Path, pathDelimiter),
	}, pathDelimiter), pathDelimiter)
}

func parseServiceHttpOption(service *protogen.Service) *HttpOption {
	option := &HttpOption{}
	if serviceHttpOption, ok := proto.GetExtension(service.Desc.Options(), httppb.E_ServiceHttp).(*httppb.Http); ok {
		if serviceHttpOption != nil && serviceHttpOption.Group != "" {
			option.Group = serviceHttpOption.Group
			option.Api = serviceHttpOption.Api
			option.Version = serviceHttpOption.Version
			option.WriterName = serviceHttpOption.WriterName
		}
	}
	return option
}

func parseMethodHttpOption(h *httppb.Http, serviceOption *HttpOption) *HttpOption {
	option := &HttpOption{}
	if h.Api == "" {
		option.Api = serviceOption.Api
	}
	if h.Version == "" {
		option.Version = serviceOption.Version
	}
	if h.Group == "" {
		option.Group = serviceOption.Group
	}
	if h.WriterName == "" {
		option.WriterName = serviceOption.WriterName
	}
	switch h.GetPattern().(type) {
	case *httppb.Http_Get:
		option.Method = http.MethodGet
		option.Path = h.GetGet()
	case *httppb.Http_Put:
		option.Method = http.MethodPut
		option.Path = h.GetPut()
		option.Body = "*"
	case *httppb.Http_Post:
		option.Method = http.MethodPost
		option.Path = h.GetPost()
		option.Body = "*"
	case *httppb.Http_Delete:
		option.Method = http.MethodDelete
		option.Path = h.GetDelete()
	case *httppb.Http_Patch:
		option.Method = http.MethodPatch
		option.Path = h.GetPatch()
		option.Body = "*"
	case *httppb.Http_Head:
		option.Method = http.MethodHead
		option.Path = h.GetHead()
	case *httppb.Http_Options:
		option.Method = http.MethodOptions
		option.Path = h.GetOptions()
	}
	return option
}

// ParseMethodHttpOption 解析method http option
// method 没有配置则使用service的
// service没有配置则解析package名称
func ParseMethodHttpOption(service *protogen.Service, h *httppb.Http) *HttpOption {
	methodOption := parseMethodHttpOption(h, parseServiceHttpOption(service))
	if methodOption.Api == "" || methodOption.Group == "" {
		serviceFullNameList := strings.Split(string(service.Desc.FullName()), ".")
		if len(serviceFullNameList) < 2 {
			panic("invalid package name")
		}
		if methodOption.Api == "" {
			methodOption.Api = serviceFullNameList[0]
		}
		if methodOption.Version == "" {
			methodOption.Version = serviceFullNameList[1]
		}
	}
	return methodOption
}
