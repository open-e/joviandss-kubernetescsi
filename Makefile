GOCMD := go
GOBUILD := $(GOCMD) build
ENV_PROD := "CGO_ENABLED=0 GOOS=linux"
IMAGE_VERSION=$(shell git describe --long --tags)
ENV_DEV := "CGO_ENABLED=1 GOOS=linux"

PROTOCOL_TYPE ?= "nfs"
SUPPORTED_PROTOCOL_TYPES := iscsi nfs

FLAGS_PROD := "-a -ldflags ' \
				-s \
				-w \
				-X github.com/open-e/joviandss-kubernetescsi/pkg/common.Version=$(IMAGE_VERSION) \
				-X github.com/open-e/joviandss-kubernetescsi/pkg/common.PluginProtocolCompileString=$(PROTOCOL_TYPE) \
				-extldflags \"-static\"'"
FLAGS_DEV := "-a -race -gcflags 'all=-N -l' -ldflags ' \
				-X github.com/open-e/joviandss-kubernetescsi/pkg/common.Version=$(IMAGE_VERSION) \
				-X github.com/open-e/joviandss-kubernetescsi/pkg/common.PluginProtocolCompileString=$(PROTOCOL_TYPE) \
				-extldflags \"-static\"'"

ENV=$(shell echo $(ENV_DEV))

FLAGS=$(shell echo $(FLAGS_DEV))

REGISTRY_NAME=opene
IMAGE_NAME=joviandss-csi
IMAGE_TAG_CENTOS=$(REGISTRY_NAME)/$(IMAGE_NAME)-c:$(IMAGE_VERSION)
IMAGE_TAG_PROD=$(REGISTRY_NAME)/$(IMAGE_NAME):$(IMAGE_VERSION)
IMAGE_TAG_LATEST_PROD=$(REGISTRY_NAME)/$(IMAGE_NAME):latest
IMAGE_TAG_DEV=$(REGISTRY_NAME)/$(IMAGE_NAME)-dev:$(IMAGE_VERSION)
IMAGE_TAG_LATEST_DEV=$(REGISTRY_NAME)/$(IMAGE_NAME)-dev:latest
IMAGE_LATEST_CENTOS=$(REGISTRY_NAME)/$(IMAGE_NAME)-c:latest
IMAGE_LATEST_UBUNTU=$(REGISTRY_NAME)/$(IMAGE_NAME)-u:latest
IMAGE_LATEST_UBUNTU_16=$(REGISTRY_NAME)/$(IMAGE_NAME)-u-16:latest



.PHONY: default all bin dev prod cli plugin container clean

default: bin


all: check-protocol-type prod

check-protocol-type:
	@echo "Checking protocol type selection..."
	@if ! echo "$(SUPPORTED_PROTOCOL_TYPES)" | grep -wq "$(PROTOCOL_TYPE)"; then \
		echo "Error: Protocol $(PROTOCOL_TYPE) is not supported. Choose from $(SUPPORTED_PROTOCOL_TYPES)"; \
		exit 1; \
	fi
	@echo "Building $(PROTOCOL_TYPE)"; \

bin:  check-protocol-type
	@$(MAKE) cli FLAGS=$(FLAGS_DEV) ENV=$(ENV_DEV)
	@$(MAKE) plugin FLAGS=$(FLAGS_DEV) ENV=$(ENV_DEV)

dev: check-protocol-type
	$(MAKE) container FLAGS=$(FLAGS_DEV)  IMAGE_TAG=$(IMAGE_TAG_DEV)  IMAGE_TAG_LATEST=$(IMAGE_TAG_LATEST_DEV)  ENV=$(ENV_DEV)

prod: check-protocol-type
	$(MAKE) container FLAGS=$(FLAGS_PROD) IMAGE_TAG=$(IMAGE_TAG_PROD) IMAGE_TAG_LATEST=$(IMAGE_TAG_LATEST_PROD) ENV=$(ENV_PROD)

cli: check-protocol-type
	$(GOCMD) mod tidy
	$(ENV) $(GOBUILD) $(FLAGS) -o _output/jdss-csi-cli ./app/jdss-csi-cli

plugin: check-protocol-type
	$(GOCMD) mod tidy
	$(ENV) $(GOBUILD) $(FLAGS) -o _output/jdss-csi-plugin ./app/joviandssplugin

container: plugin cli
	@echo Building Container $(IMAGE_TAG) $(IMAGE_TAG_LATEST)
	podman build -t docker.io/$(IMAGE_TAG) -f ./deploy/container/centos.Dockerfile .
	podman build -t docker.io/$(IMAGE_TAG_LATEST) -f ./deploy/container/centos.Dockerfile .

clean:
	go clean -r -x
