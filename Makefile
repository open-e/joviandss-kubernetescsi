REGISTRY_NAME=opene
IMAGE_NAME=joviandss-csi
IMAGE_VERSION=$(shell git describe --long --tags)
IMAGE_TAG_CENTOS=$(REGISTRY_NAME)/$(IMAGE_NAME)-c:$(IMAGE_VERSION)
IMAGE_TAG_UBUNTU=$(REGISTRY_NAME)/$(IMAGE_NAME)-u:$(IMAGE_VERSION)
IMAGE_TAG_UBUNTU_16=$(REGISTRY_NAME)/$(IMAGE_NAME)-u-16:$(IMAGE_VERSION)
IMAGE_LATEST_CENTOS=$(REGISTRY_NAME)/$(IMAGE_NAME)-c:latest
IMAGE_LATEST_UBUNTU=$(REGISTRY_NAME)/$(IMAGE_NAME)-u:latest
IMAGE_LATEST_UBUNTU_16=$(REGISTRY_NAME)/$(IMAGE_NAME)-u-16:latest

.PHONY: default all joviandss clean hostpath-container iscsi rest

default: joviandss



all:  joviandss joviandss-container

joviandss:
	go get ./pkg/joviandss
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-X JovianDSS-KubernetesCSI/pkg/joviandss.Version=$(IMAGE_VERSION) -extldflags "-static"' -o _output/jdss-csi-plugin ./app/joviandssplugin

joviandss-container: joviandss
	@echo Building Container
	sudo docker build -t $(IMAGE_TAG_CENTOS) -f ./app/joviandssplugin/centos.Dockerfile .
	sudo docker build -t $(IMAGE_TAG_UBUNTU) -f ./app/joviandssplugin/ubuntu.Dockerfile .
	sudo docker build -t $(IMAGE_TAG_UBUNTU_16) -f ./app/joviandssplugin/ubuntu-16.Dockerfile .

clean:
	go clean -r -x
	-rm -rf _outpusudo docker push $(IMAGE_TAG_CENTOS)
	-rm -rf _outpusudo docker push $(IMAGE_TAG_UBUNTU)
	-rm -rf _outpusudo docker push $(IMAGE_TAG_UBUNTU_16)
