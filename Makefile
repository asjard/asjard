export BIFROST_DIR ?= ./third_party/bifrost

.PHONY: test

-include $(BIFROST_DIR)/Makefile_base

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

build_gen_go_validate: ## 生成protoc-gen-go-validate命令
	go build -o $(GOPATH)/bin/protoc-gen-go-validate -ldflags '-w -s' ./cmd/protoc-gen-go-validate/*.go

build_gen_go_rest2grpc_gw: ## 生成protoc-gen-go-rest2grpc-gw命令
	go build -o $(GOPATH)/bin/protoc-gen-go-rest2grpc-gw -ldflags '-w -s' ./cmd/protoc-gen-go-rest2grpc-gw/*.go

build_gen_ts: ## 生成protoc-gen-ts命令
	go build -o $(GOPATH)/bin/protoc-gen-ts -ldflags '-w -s' ./cmd/protoc-gen-ts/*.go

github_workflows_dependices: docker-compose.yaml ## github workflows 依赖环境
	docker compose -p asjard up -d

github_workflows_test: update github_workflows_dependices test ## github workflow 运行测试用例

test: gocyclo govet ## 运行测试用例
	go test -race -cover -coverprofile=cover.out $$(go list ./...|grep -v cmd|grep -v 'protobuf/')
	# go tool cover -html=cover.out

gocyclo: ## 圈复杂度检测
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	gocyclo -over 15 -ignore third_party/ .

govet: ## 静态检查
	go vet -all ./...

zed_clean: ## zed编辑器清理
	for file in $$(find . -name '._*'); \
	do \
	   rm -rf $$file; \
	done
