export BIFROST_DIR ?= ./third_party/bifrost


-include $(BIFROST_DIR)/Makefile_base

.PHONY: update
update: .gitmodules ## Update submodule
	git submodule sync
	git submodule foreach --recursive git reset --hard
	git submodule foreach --recursive git clean -fdx
	git submodule init
	git submodule update
	git submodule update --remote
	git submodule foreach  --recursive 'tag="$$(git config -f $$toplevel/.gitmodules submodule.$$name.tag)";[ -n $$tag ] && git reset --hard  $$tag || echo "this module has no tag"'

.PHONY: build_cipher_aes
build_cipher_aes: ## Build command asjard_cipher_aes
	go build -o $(GOPATH)/bin/asjard_cipher_aes -ldflags '-w -s' ./cmd/asjard_cipher_aes/*.go

.PHONY: build_gen_go_rest
build_gen_go_rest: ## Build command protoc-gen-go-rest
	go build -o $(GOPATH)/bin/protoc-gen-go-rest -ldflags '-w -s' ./cmd/protoc-gen-go-rest/*.go

.PHONY: build_gen_go_validate
build_gen_go_validate: ## Build command protoc-gen-go-validate
	go build -o $(GOPATH)/bin/protoc-gen-go-validate -ldflags '-w -s' ./cmd/protoc-gen-go-validate/*.go

.PHONY: build_gen_go_asynq
build_gen_go_asynq: ## Build command protoc-gen-go-rest
	go build -o $(GOPATH)/bin/protoc-gen-go-asynq -ldflags '-w -s' ./cmd/protoc-gen-go-asynq/*.go

.PHONY: build_gen_go_amqp
build_gen_go_amqp: ## build command protoc-gen-go-amqp
	go build -o $(GOPATH)/bin/protoc-gen-go-amqp -ldflags '-w -s' ./cmd/protoc-gen-go-amqp/*.go

.PHONY: build_gen_go_rest2grpc_gw
build_gen_go_rest2grpc_gw: ## Build command protoc-gen-go-rest2grpc-gw
	go build -o $(GOPATH)/bin/protoc-gen-go-rest2grpc-gw -ldflags '-w -s' ./cmd/protoc-gen-go-rest2grpc-gw/*.go

.PHONY: build_gen_ts
build_gen_ts: ## Build command protoc-gen-ts
	go build -o $(GOPATH)/bin/protoc-gen-ts -ldflags '-w -s' ./cmd/protoc-gen-ts/*.go

.PHONY: build_gen_ts_enum
build_gen_ts_enum: ## Build command protoc-gen-ts-enum
	go build -o $(GOPATH)/bin/protoc-gen-ts-enum -ldflags '-w -s' ./cmd/protoc-gen-ts-enum/*.go

.PHONY: build_gen_ts_umi
build_gen_ts_umi: ## Build command protoc-gen-ts-umi
	go build -o $(GOPATH)/bin/protoc-gen-ts-umi -ldflags '-w -s' ./cmd/protoc-gen-ts-umi/*.go

.PHONY: gen_proto
gen_proto: clean ## Build protobuf
	bash third_party/github.com/asjard/protobuf/build.sh

.PHONY: github_workflows_dependices
github_workflows_dependices: docker-compose.yaml ## Install github workflows environment
	docker compose -p asjard up -d

.PHONY: github_workflows_test
github_workflows_test: update github_workflows_dependices test ## Run unit test in github workflow

.PHONY: test
test: clean gocyclo govet ## Run unit test
	go test -race -cover -coverprofile=cover.out $$(go list ./...|grep -v cmd|grep -v 'protobuf/')
	go test -benchmem -bench=. -run=^$$ $$(go list ./...|grep -v cmd|grep -v 'protobuf/')

	# go tool cover -html=cover.out

.PHONY: gocyclo
gocyclo: ## Cyclo check
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	gocyclo -over 15 -ignore third_party/ .

.PHONY: govet
govet: ## Static check
	go vet -all ./...

.PHONY: clean
clean: ## Clean
	# go clean -cache -testcache -modcache
	find . -name '._*' -delete
