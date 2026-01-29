export BIFROST_DIR ?= ../third_party/bifrost
export PROTO_DIR ?= ./protos-repo
export PROTOC_OPT ?= "-I../third_party"


export GEN_PROTO_GO_VALIDATE ?= true
export GEN_PROTO_GO_REST ?= true
export GEN_PROTO_GO_REST_GW ?= true
export GEN_PROTO_GO_AMQP ?= true
export GEN_PROTO_GO_VALIDATE_OPT ?= validate_enum=true
# export DEBUG ?= true

-include $(BIFROST_DIR)/Makefile_build

services ?= svc-example/apis/api \
			svc-gw/apis/api

.PHONY: run_dev
run_dev: ## 本地运行服务
	ASJARD_CONF_DIR="$(PWD)/$(services)/conf $(PWD)/$(services)/../../conf $(PWD)/conf"  go run -race ./$(services)/
