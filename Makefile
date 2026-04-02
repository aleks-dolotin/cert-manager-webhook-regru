IMAGE ?= ghcr.io/aleks-dolotin/cert-manager-webhook-regru
VERSION ?= $(shell cat VERSION 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags="-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)"

.PHONY: build test lint docker-build bump-version test-smoke

build:
	go build $(LDFLAGS) -o bin/webhook ./cmd/webhook

build-cli:
	go build $(LDFLAGS) -o bin/regru-cli ./cmd/cli

test:
	go test ./... -v

test-race:
	go test ./... -race -v

lint:
	golangci-lint run

check: test test-race lint

# Smoke test against real Reg.ru API.
# Requires: REGRU_USERNAME, REGRU_PASSWORD, SMOKE_TEST_ZONE
test-smoke:
	@echo "Running smoke tests against real Reg.ru API..."
	@echo "Required env: REGRU_USERNAME, REGRU_PASSWORD, SMOKE_TEST_ZONE"
	go test -v -tags=smoke -count=1 -timeout=300s ./tests/smoke/...

docker-build:
	docker build -t $(IMAGE):$(VERSION) --build-arg VERSION=$(VERSION) .

docker-push: docker-build
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest

# Bump version in VERSION file, Chart.yaml, and values.yaml.
# Auto-increments minor version by default.
# Override: make bump-version VERSION_NEW=1.2.0
bump-version:
	@OLD=$$(cat VERSION); \
	if [ -n "$(VERSION_NEW)" ]; then \
		NEW="$(VERSION_NEW)"; \
	else \
		MAJOR=$$(echo $$OLD | cut -d. -f1); \
		MINOR=$$(echo $$OLD | cut -d. -f2); \
		PATCH=$$(echo $$OLD | cut -d. -f3); \
		NEW="$$MAJOR.$$((MINOR + 1)).$$PATCH"; \
	fi; \
	echo "Bumping $$OLD → $$NEW"; \
	echo "$$NEW" > VERSION; \
	sed -i.bak \
		-e "s/^version: [0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*$$/version: $$NEW/" \
		-e "s/^appVersion: \"[0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\"$$/appVersion: \"$$NEW\"/" \
		charts/cert-manager-webhook-regru/Chart.yaml && rm -f charts/cert-manager-webhook-regru/Chart.yaml.bak; \
	sed -i.bak \
		-e "s/tag: [a-z0-9][a-z0-9.-]*/tag: v$$NEW/" \
		charts/cert-manager-webhook-regru/values.yaml && rm -f charts/cert-manager-webhook-regru/values.yaml.bak; \
	echo "Done. Review: git diff"
