# Set arguments and default values
ARG GO_VERSION=1.16
ARG BUILDER_IMAGE=golang:${GO_VERSION}
ARG BASE_IMAGE=alpine:3.14
ARG LISTEN_PORT=8000

#### STEP 1- Build image ####
FROM ${BUILDER_IMAGE} as builder
# Set env variables
ENV GOOS=linux \
    GOARCH=amd64

# ARG DEVOPS_TOKEN and ENV GOPRIVATE are used in order to import Tivan's instrumenation from private module in AzureDevOps.
# More detailes on why we need those varaible can be found in https://seb-nyberg.medium.com/using-go-modules-with-private-azure-devops-repositories-4664b621f782
ARG DEVOPS_TOKEN
# Tivan use dependency server for GO - https://dev.azure.com/msazure/One/_git/Rome-Detection-Tivan-GoDependenciesResolver
ENV GOPRIVATE "tivan.ms"
# Config git to wore with azuredevops and use DEVOPS_TOKEN token
RUN git config --global url."https://azure-defender:$DEVOPS_TOKEN@dev.azure.com".insteadOf "https://dev.azure.com"

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
# Run the executable file - azdproxy
ENTRYPOINT ["./azdproxy"]