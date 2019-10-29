FROM golang:1.13.1-alpine3.10 AS build
RUN apk add --no-cache curl git build-base
WORKDIR $GOPATH/src/github.com/go-ocf/openapi-connector
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /go/bin/service ./cmd/service

FROM alpine:3.10 as service
RUN apk add --no-cache ca-certificates
COPY --from=build /go/bin/service /usr/local/bin/service
ENTRYPOINT ["/usr/local/bin/service"]