all: help

help: ## 使用帮助
	@echo "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\033[36m\1\\033[m:\2/' | column -c2 -t -s :)"

gen_example_proto: ## 生成examples目录下的协议
	/bin/bash scripts/gen_example_proto.sh

build_gen_go_rest: ## 生成protoc-gen-go-rest命令
	go build -o $(GOPATH)/bin/protoc-gen-go-rest ./cmd/protoc-gen-go-rest/*.go
