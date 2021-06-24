#### STEP 1- Build image ####
FROM golang:1.16 as builder

# Set env variables
ENV GOOS=linux \
    GOARCH=amd64 \
    baseDir=/go

# Create and set working directory
RUN mkdir -p $baseDir/src/AzureDefender-K8S-InClusterDefense/src
WORKDIR $baseDir/src/AzureDefender-K8S-InClusterDefense/src
# Copy dependancies files(mod/sum) and download them - will also be cached if we won't change mod/sum
# https://petomalina.medium.com/using-go-mod-download-to-speed-up-golang-docker-builds-707591336888
COPY go.mod .
COPY go.sum .
RUN go mod download
# Copy all code to workdir
COPY . .

# build binary files
RUN	go build -o $baseDir/main .

#### STEP 2 - Build base image ####
# alpine makes our image build roughly ~40MB
FROM alpine

ENV baseDir /go
# set working directory
WORKDIR ${baseDir}

# Add dependencies for base image
# Support Golang binaries in alpine:
# See https://stackoverflow.com/questions/36279253/go-compiled-binary-wont-run-in-an-alpine-docker-container-on-ubuntu-host
# Also add Bash for interactive shell
RUN apk add --no-cache \
        libc6-compat \
        bash

# Copy build artifacts from builder
COPY --from=builder ${baseDir}/main ./azdproxy
# expose port for webhook server
EXPOSE 8000
# run the binary
ENTRYPOINT ["./azdproxy"]