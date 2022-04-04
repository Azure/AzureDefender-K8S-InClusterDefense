# Set arguments and default values
ARG GO_VERSION=1.16
ARG BUILDER_IMAGE=golang:${GO_VERSION}
ARG BASE_IMAGE=alpine:3.15.3
ARG LISTEN_PORT=8000

#### STEP 1- Build image ####
FROM ${BUILDER_IMAGE} as builder
# Set env variables
ENV GOOS=linux \
    GOARCH=amd64

# Create and set working directory - directories have to be in the match path of the packages in order to import them.
RUN mkdir -p /go/src/github.com/Azure/AzureDefender-K8S-InClusterDefense
WORKDIR /go/src/github.com/Azure/AzureDefender-K8S-InClusterDefense
# Copy dependancies files(mod/sum) and download them - will also be cached if we won't change mod/sum
# https://petomalina.medium.com/using-go-mod-download-to-speed-up-golang-docker-builds-707591336888
COPY go.mod .
COPY go.sum .

RUN go mod download
# Copy all go files when finish downloading all dependecies
COPY pkg/ pkg/
COPY cmd/ cmd/
COPY main.go main.go
# Build binary files
# In order to use net library, we have to disable CGO (CGO_ENABLED = 0)
# See https://stackoverflow.com/questions/36279253/go-compiled-binary-wont-run-in-an-alpine-docker-container-on-ubuntu-host
RUN	CGO_ENABLED=0 go build -o /go/azdproxy .

#### STEP 2 - Build base image ####
FROM ${BASE_IMAGE}
# Copy executable file into the new container
COPY --from=builder /go/azdproxy ./azdproxy

# Expose port for webhook server
EXPOSE ${LISTEN_PORT}
# Add label of maintainers:
MAINTAINER ascdetectiontomer@microsoft.com

# Run the executable file - azdproxy
ENTRYPOINT ["./azdproxy"]