IMAGE ?= ghcr.io/aleks-dolotin/cert-manager-webhook-regru
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

.PHONY: build test lint docker-build

build:
	go build -o bin/webhook ./cmd/webhook

test:
	go test ./... -v

test-race:
	go test ./... -race -v

lint:
	golangci-lint run

check: test test-race lint

docker-build:
	docker build -t $(IMAGE):$(VERSION) --build-arg VERSION=$(VERSION) .

docker-push: docker-build
	docker push $(IMAGE):$(VERSION)
