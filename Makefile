.PHONY: genmocks
genmocks: ${GOPATH}/bin/mockgen
	mockgen -source=handler.go -destination=./mocks/handler.go -typed=true

GOPATH ?= $(shell go env GOPATH)
MOCKGEN_VERSION := v0.4.0
${GOPATH}/bin/mockgen:
	go install go.uber.org/mock/mockgen@${MOCKGEN_VERSION}
	
.PHONY: cleanmocks
cleanmocks:
	find . -type d -name "mocks" | xargs -I{} rm -rf {}

.PHONY: deps
deps: ${GOPATH}/bin/mockgen
	go mod download

.PHONY: test
test: deps genmocks
	go test -v -race ./...

GOLANGCI_LINT_VERSION := v1.58.0

.PHONY: lint
lint: ${GOPATH}/bin/golangci-lint
	golangci-lint run --max-same-issues=999 --max-issues-per-linter=999 --config=./.golangci-lint.yaml
	
${GOPATH}/bin/golangci-lint:	
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}

