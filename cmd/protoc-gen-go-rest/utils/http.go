package utils

import (
	"net/http"
	"strings"

	"github.com/asjard/asjard/pkg/protobuf/httppb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func ParseMethodOption(service *protogen.Service, httpOption *httppb.Http) (path, method, body string) {
	apiGroup := ""
	apiClassify := ""
	apiVersion := ""
	if serviceHttpOption, ok := proto.GetExtension(service.Desc.Options(), httppb.E_ServiceHttp).(*httppb.Http); ok {
		if serviceHttpOption != nil && serviceHttpOption.Group != "" {
			apiGroup = serviceHttpOption.Group
			apiClassify = serviceHttpOption.Api
			apiVersion = serviceHttpOption.Version
		}
	}
	switch httpOption.GetPattern().(type) {
	case *httppb.Http_Get:
		method = http.MethodGet
		path = httpOption.GetGet()
	case *httppb.Http_Put:
		method = http.MethodPut
		path = httpOption.GetPut()
		body = "*"
	case *httppb.Http_Post:
		method = http.MethodPost
		path = httpOption.GetPost()
		body = "*"
	case *httppb.Http_Delete:
		method = http.MethodDelete
		path = httpOption.GetDelete()
	case *httppb.Http_Patch:
		method = http.MethodPatch
		path = httpOption.GetPatch()
		body = "*"
	case *httppb.Http_Head:
		method = http.MethodHead
		path = httpOption.GetHead()
	}
	// 根据package名称解析
	// api.v1.xxx
	// 第一部分为接口类型
	// 第二部分为接口版本
	apiClassify = httpOption.Api
	apiVersion = httpOption.Version
	if apiClassify == "" || apiVersion == "" {
		serviceFullNameList := strings.Split(string(service.Desc.FullName()), ".")
		if len(serviceFullNameList) < 2 {
			panic("invalid package name")
		}
		if apiClassify == "" {
			apiClassify = serviceFullNameList[0]
		}
		if apiVersion == "" {
			apiVersion = serviceFullNameList[1]
		}
	}
	if httpOption.Group != "" {
		apiGroup = httpOption.Group
	}
	fullPath := ""
	apiClassify = strings.Trim(apiClassify, "/")
	if apiClassify != "" {
		fullPath += "/" + apiClassify
	}
	apiVersion = strings.Trim(apiVersion, "/")
	if apiVersion != "" {
		fullPath += "/" + apiVersion
	}
	apiGroup = strings.Trim(apiGroup, "/")
	if apiGroup != "" {
		fullPath += "/" + apiGroup
	}
	path = strings.Trim(path, "/")
	if path != "" {
		fullPath += "/" + path
	}
	if fullPath == "" {
		fullPath = "/"
	}
	path = fullPath
	return
}
