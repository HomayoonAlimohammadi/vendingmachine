GOPATH ?= $(shell go env GOPATH)
MOCKGEN_VERSION := v0.4.0
GOLANGCI_LINT_VERSION := v1.58.0
BUILD_IMAGE_NAME ?= "vendingmachine:1.0.0"
# these should be changed according to the
# config.yaml file
CONTAINER_PORT := 8080
HOST_PORT := 8080

.PHONY: genmocks
genmocks: ${GOPATH}/bin/mockgen
	mockgen -source=handler.go -destination=./mocks/handler.go -typed=true

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

.PHONY: lint
lint: ${GOPATH}/bin/golangci-lint deps genmocks
	golangci-lint run --max-same-issues=999 --max-issues-per-linter=999 --config=./.golangci-lint.yaml
	
${GOPATH}/bin/golangci-lint:	
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}

.PHONY: build
build:
	go build -o vendingmachine .

.PHONY: build-image
build-image:
	docker build ${BUILD_IMAGE_NAME} .

.PHONY: push
push-image:
	docker push ${BUILD_IMAGE_NAME}

.PHONY: run
run: build
	./vendingmachine

.PHONY:
run-container: build
	docker run -p 8080:8080 ${BUILD_IMAGE_NAME}