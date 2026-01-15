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

build_gen_go_asynq: ## 生成protoc-gen-go-rest命令
	go build -o $(GOPATH)/bin/protoc-gen-go-asynq -ldflags '-w -s' ./cmd/protoc-gen-go-asynq/*.go

build_gen_go_amqp: ## 生成protoc-gen-go-amqp命令
	go build -o $(GOPATH)/bin/protoc-gen-go-amqp -ldflags '-w -s' ./cmd/protoc-gen-go-amqp/*.go

build_gen_go_rest2grpc_gw: ## 生成protoc-gen-go-rest2grpc-gw命令
	go build -o $(GOPATH)/bin/protoc-gen-go-rest2grpc-gw -ldflags '-w -s' ./cmd/protoc-gen-go-rest2grpc-gw/*.go

build_gen_ts: ## 生成protoc-gen-ts命令
	go build -o $(GOPATH)/bin/protoc-gen-ts -ldflags '-w -s' ./cmd/protoc-gen-ts/*.go

build_gen_ts_enum: ## 生成protoc-gen-ts-enum命令
	go build -o $(GOPATH)/bin/protoc-gen-ts-enum -ldflags '-w -s' ./cmd/protoc-gen-ts-enum/*.go

build_gen_ts_umi: ## 生成protoc-gen-ts-umi命令
	go build -o $(GOPATH)/bin/protoc-gen-ts-umi -ldflags '-w -s' ./cmd/protoc-gen-ts-umi/*.go

gen_proto: clean ## 生成协议文件
	bash third_party/github.com/asjard/protobuf/build.sh

github_workflows_dependices: docker-compose.yaml ## github workflows 依赖环境
	docker compose -p asjard up -d

github_workflows_test: update github_workflows_dependices test ## github workflow 运行测试用例

test: clean gocyclo govet ## 运行测试用例
	go test -race -cover -coverprofile=cover.out $$(go list ./...|grep -v cmd|grep -v 'protobuf/')
	go test -benchmem -bench=. -run=^$$ $$(go list ./...|grep -v cmd|grep -v 'protobuf/')

	# go tool cover -html=cover.out

gocyclo: ## 圈复杂度检测
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	gocyclo -over 15 -ignore third_party/ .

govet: ## 静态检查
	go vet -all ./...

clean: ## 清理
	find . -name '._*' -delete
