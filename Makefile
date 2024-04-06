.DEFAULT_GOAL=help

SERVICE_NAME=go-sandbox-service
SERVICE_PORT?=8080

GO_VERSION?=1.22.0
GO_ARCH?=linux-amd64

AARCH:=$(shell uname -m)
ifeq ($(findstring arm,$(AARCH)),arm)
  GO_ARCH=linux-arm64
endif

generate: ## compile proto files for Go
	./scripts/gen_proto.sh

build: generate ## build docker image
	docker build --build-arg=GO_VERSION=$(GO_VERSION) \
		--build-arg=GO_ARCH=$(GO_ARCH) \
		--build-arg=SERVICE_PORT=$(SERVICE_PORT) \
		-t $(SERVICE_NAME) .

run: ## runs sandbox as a gRPC service
	# isolate needs access to create/modify cgroups for sandboxing which requires privileged access
	docker run --privileged --rm --name $(SERVICE_NAME)  -p $(SERVICE_PORT):$(SERVICE_PORT) $(SERVICE_NAME)

help:
	@grep -E '^[a-zA-Z_-]+:.*?\#\# .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?\#\# "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
