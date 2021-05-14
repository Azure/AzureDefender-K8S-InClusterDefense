FROM golang:1.13.4-alpine as builder
RUN apk add --update make
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go
COPY . /go/src/github.com/Azure/opa-asc-proxy
WORKDIR /go/src/github.com/Azure/opa-asc-proxy
ARG IMAGE_VERSION=0.0.1
RUN make build

FROM alpine:3.10.3
RUN apk update && apk add --no-cache \
    bash \
    curl \
    jq
COPY --from=builder /go/src/github.com/Azure/opa-asc-proxy/opa-asc-proxy /bin/
COPY getimagesha.sh /
RUN chmod a+x /bin/opa-asc-proxy
RUN chmod a+x /getimagesha.sh

ENTRYPOINT ["/bin/opa-asc-proxy"]
