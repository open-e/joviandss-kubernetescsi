REGISTRY_NAME=opene
IMAGE_NAME=joviandss-csi
IMAGE_VERSION=$(shell git describe --long --tags)
IMAGE_TAG_CENTOS=$(REGISTRY_NAME)/$(IMAGE_NAME)-c:$(IMAGE_VERSION)
IMAGE_TAG=$(REGISTRY_NAME)/$(IMAGE_NAME):$(IMAGE_VERSION)
IMAGE_TAG_LATEST=$(REGISTRY_NAME)/$(IMAGE_NAME):latest
IMAGE_TAG_DEV=$(REGISTRY_NAME)/$(IMAGE_NAME)-dev:$(IMAGE_VERSION)
IMAGE_TAG_DEV_LATEST=$(REGISTRY_NAME)/$(IMAGE_NAME)-dev:latest
#IMAGE_TAG_UBUNTU_16=$(REGISTRY_NAME)/$(IMAGE_NAME)-u-16:$(IMAGE_VERSION)
IMAGE_LATEST_CENTOS=$(REGISTRY_NAME)/$(IMAGE_NAME)-c:latest
IMAGE_LATEST_UBUNTU=$(REGISTRY_NAME)/$(IMAGE_NAME)-u:latest
IMAGE_LATEST_UBUNTU_16=$(REGISTRY_NAME)/$(IMAGE_NAME)-u-16:latest

.PHONY: default all joviandss clean hostpath-container iscsi rest

default: joviandss



all:  joviandss container cli

cli:
	go mod tidy
	go get ./app/joviandssplugin
	CGO_ENABLED=0 GOOS=linux go build -a -o _output/jdss-csi-cli ./app/jdss-csi-cli

joviandss:
	go mod tidy
	go get ./app/joviandssplugin
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-X joviandss-kubernetescsi/pkg/common.Version=$(IMAGE_VERSION) -extldflags "-static"' -o _output/jdss-csi-plugin ./app/joviandssplugin
	#chmod +x _output/jdss-csi-plugin

container: joviandss cli
	@echo Building Container
	podman build -t docker.io/$(IMAGE_TAG) -f ./deploy/container/centos.Dockerfile .
	podman build -t docker.io/$(IMAGE_TAG_LATEST) -f ./deploy/container/centos.Dockerfile .

containerdev: joviandss cli
	@echo Building Container
	podman build -t docker.io/$(IMAGE_TAG_DEV) -f ./deploy/container/centos.Dockerfile .
	podman build -t docker.io/$(IMAGE_TAG_DEV_LATEST) -f ./deploy/container/centos.Dockerfile .

clean:
	go clean -r -x
	-rm -rf _outpusudo podman push $(IMAGE_TAG_CENTOS)
	-rm -rf _outpusudo podman push $(IMAGE_TAG_UBUNTU)
	#-rm -rf _outpusudo podman push $(IMAGE_TAG_UBUNTU_16)
