# Builder Image
FROM golang:1.16 as builder

# create and set working directory
RUN mkdir -p /app
COPY tag_2_digest.sh /app
RUN chmod a+x /app/tag_2_digest.sh
RUN sed -i -e 's/\r//g' /app/tag_2_digest.sh

WORKDIR /app
# install dependencies
ADD go.mod go.mod
ADD go.sum go.sum
RUN go mod download
# add code
ADD . .
# build the source
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main

# Final Image
FROM alpine:3.12
# add ca-certificates
RUN apk update && apk --no-cache  add ca-certificates \
    bash \
    curl \
    jq
# set working directory
WORKDIR /app
# copy the binary from builder
COPY --from=builder /app/main ./main
# run the binary
CMD ["./main"]