package bootstrap

import (
	// 服务端拦截器
	_ "github.com/asjard/asjard/pkg/server/interceptors"
	// 客户端拦截器
	_ "github.com/asjard/asjard/pkg/client/interceptors"
)
