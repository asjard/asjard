all: help

##env 是否根据protobuf文件生成*.pb.go文件
GEN_PROTO_GO ?= true
##env 是否根据protobuf文件生成*_grpc.pb.go文件
GEN_PROTO_GO_GRPC ?= true
##env 是否根据protobuf文件生成*_rest.pb.go文件
GEN_PROTO_GO_REST ?= true
##env 是否根据protobuf文件生成*_rest_gw.pb.go文件
GEN_PROTO_GO_REST_GW ?= true
##env 是否根据protobuf文件生成*_ts.pb.go文件
GEN_PROTO_TS ?= false

.PHONY: test

help: ## 使用帮助
	@echo "Commands:"
	@echo "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/  \\033[35m\1\\033[m:\2/' | column -c2 -t -s :)"
	@echo
	@echo "Envs:"
	@echo "$$(grep -A1 -hE '^##env' Makefile |sed 's/##env//' |sed 'N;s/\n/ =/'|awk -F '=' '{print $$2"|"$$1"default:"$$3}'|sed -e 's/ .*|/|/' -e 's/^\(.\+\)|\(.*\)/  \\033[32m\1\\033[m|\2/'  |column -c2 -t -s '|')"

update: .gitmodules ## 更新本地代码
	git submodule sync
	git submodule foreach --recursive git reset --hard
	git submodule foreach --recursive git clean -fdx
	git submodule init
	git submodule update
	git submodule update --remote
	git submodule foreach  --recursive 'tag="$$(git config -f $$toplevel/.gitmodules submodule.$$name.tag)";[ -n $$tag ] && git reset --hard  $$tag || echo "this module has no tag"'

build_cipher_aes: ## 生成asjard_cipher_aes命令
	go build -o $(GOPATH)/bin/asjard_cipher_aes -ldflags '-w -s' ./cmd/asjard_cipher_aes/*.go

build_gen_go_rest: ## 生成protoc-gen-go-rest命令
	go build -o $(GOPATH)/bin/protoc-gen-go-rest -ldflags '-w -s' ./cmd/protoc-gen-go-rest/*.go

build_gen_go_rest2grpc_gw: ## 生成protoc-gen-go-rest2grpc-gw命令
	go build -o $(GOPATH)/bin/protoc-gen-go-rest2grpc-gw -ldflags '-w -s' ./cmd/protoc-gen-go-rest2grpc-gw/*.go

build_gen_ts: ## 生成protoc-gen-ts命令
	go build -o $(GOPATH)/bin/protoc-gen-ts -ldflags '-w -s' ./cmd/protoc-gen-ts/*.go

github_workflows_dependices: docker-compose.yaml ## github workflows 依赖环境
	docker compose -p asjard up -d

github_workflows_test: github_workflows_dependices test ## github workflow 运行测试用例

test: ## 运行测试用例
	go test -cover -coverprofile=cover.out $$(go list ./...|grep -v cmd)
	# go tool cover -html=cover.out
