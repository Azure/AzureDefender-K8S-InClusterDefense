IMAGE_VERSION ?= 0.0.2
BUILD_DATE=$$(date +%Y-%m-%d-%H:%M)
GO_FILES=$(shell go list ./...)
REGISTRY ?= ritazh
DOCKER_IMAGE ?= $(REGISTRY)/opa-asc-proxy

GO111MODULE ?= on
export GO111MODULE

ifeq ($(OS),Windows_NT)
	GOOS_FLAG = windows
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S), Linux)
		GOOS_FLAG = linux
	endif
	ifeq ($(UNAME_S), Darwin)
		GOOS_FLAG = darwin
	endif
endif

.PHONY: build
build: setup
	@echo "Building..."
	$Q GOOS=${GOOS_FLAG} CGO_ENABLED=0 go build -ldflags "-X main.BuildDate=$(BUILD_DATE) -X main.BuildVersion=$(IMAGE_VERSION)" . 

image:
# build inside docker container
	@echo "Building docker image..."
	$Q docker build --no-cache -t $(DOCKER_IMAGE):$(IMAGE_VERSION) --build-arg IMAGE_VERSION="$(IMAGE_VERSION)" .

setup:
	@echo "Setup..."
	$Q go env
