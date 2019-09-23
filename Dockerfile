FROM golang:1.13.0-alpine3.10 AS build
RUN apk add --no-cache curl git build-base && \
	curl -SL -o /usr/bin/dep https://github.com/golang/dep/releases/download/v0.5.0/dep-linux-amd64 && \
	chmod +x /usr/bin/dep
WORKDIR $GOPATH/src/github.com/go-ocf/openapi-connector
COPY . .

RUN dep ensure -v --vendor-only
RUN go build -o /go/bin/service ./cmd/service

FROM alpine:3.10 as service
RUN apk add --no-cache ca-certificates
COPY --from=build /go/bin/service /usr/local/bin/service
ENTRYPOINT ["/usr/local/bin/service"]