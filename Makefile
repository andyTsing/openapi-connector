SERVICE_NAME = $(notdir $(CURDIR))
LATEST_TAG = vnext
VERSION_TAG = vnext-$(shell git rev-parse --short=7 --verify HEAD)

default: build

define build-docker-image
	docker build \
		--network=host \
		--tag ocfcloud/$(SERVICE_NAME):$(VERSION_TAG) \
		--tag ocfcloud/$(SERVICE_NAME):$(LATEST_TAG) \
		--target $(1) \
		.
endef

build-testcontainer:
	$(call build-docker-image,build)

build-servicecontainer:
	$(call build-docker-image,service)

build: build-testcontainer build-servicecontainer

test: clean build-testcontainer
	docker-compose pull
	docker-compose up -d
	docker run \
		--network=host \
		--mount type=bind,source="$(shell pwd)",target=/shared \
		ocfcloud/$(SERVICE_NAME):$(VERSION_TAG) \
		go test -v ./... -covermode=atomic -coverprofile=/shared/coverage.txt

push: build-servicecontainer
	docker push ocfcloud/$(SERVICE_NAME):$(VERSION_TAG)
	docker push ocfcloud/$(SERVICE_NAME):$(LATEST_TAG)

clean:
	docker-compose down --volumes || true

proto/generate:
	protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf --gogofaster_out=$GOPATH/src *.proto

.PHONY: build-testcontainer build-servicecontainer build test push clean proto/generate



