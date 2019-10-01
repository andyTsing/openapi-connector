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

make-ca:
	docker pull smallstep/step-ca
	mkdir -p ./test/step-ca/data/secrets
	echo "password" > ./test/step-ca/data/secrets/password
	docker run \
		-it \
		-v "$(shell pwd)"/test/step-ca/data:/home/step --user $(shell id -u):$(shell id -g) \
		smallstep/step-ca \
		/bin/bash -c "step ca init -dns localhost -address=:10443 -provisioner=test@localhost -name test -password-file ./secrets/password && step ca provisioner add acme --type ACME"
	docker run \
		-d \
		--network=host \
		--name=step-ca-test \
		-v /etc/nsswitch.conf:/etc/nsswitch.conf \
		-v "$(shell pwd)"/test/step-ca/data:/home/step --user $(shell id -u):$(shell id -g) \
		smallstep/step-ca

test: clean make-ca build-testcontainer
	docker-compose pull
	docker-compose up -d
	docker run \
		--network=host \
		-v $(shell pwd)/test/step-ca/data/certs/root_ca.crt:/root_ca.crt \
		-e DIAL_ACME_CA_POOL=/root_ca.crt \
		-e DIAL_ACME_DOMAINS="localhost" \
		-e DIAL_ACME_DIRECTORY_URL="https://localhost:10443/acme/acme/directory" \
		-e LISTEN_ACME_CA_POOL=/root_ca.crt \
		-e LISTEN_ACME_DOMAINS="localhost" \
		-e LISTEN_ACME_DIRECTORY_URL="https://localhost:10443/acme/acme/directory" \
		--mount type=bind,source="$(shell pwd)",target=/shared \
		ocfcloud/$(SERVICE_NAME):$(VERSION_TAG) \
		go test -v ./... -covermode=atomic -coverprofile=/shared/coverage.txt

push: build-servicecontainer
	docker push ocfcloud/$(SERVICE_NAME):$(VERSION_TAG)
	docker push ocfcloud/$(SERVICE_NAME):$(LATEST_TAG)


clean:
	docker-compose down --volumes || true
	docker rm -f step-ca-test || true
	rm -rf ./test/step-ca  || true

proto/generate:
	protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf --gogofaster_out=$GOPATH/src *.proto

.PHONY: build-testcontainer build-servicecontainer build test push clean proto/generate make-ca



