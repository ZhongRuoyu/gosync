export GOBIN ?= $(shell pwd)/bin
GOLINT = $(GOBIN)/golint
STATICCHECK = $(GOBIN)/staticcheck

GO_FILES = $(shell find . -name '*.go' -not -path './vendor/*' -print)


.PHONY: all
all: build examples

.PHONY: build
build:
	go build ./...

.PHONY: examples
examples:
	mkdir -p bin
	cd bin && go build ../examples/...


.PHONY: lint
lint: gofmt golint staticcheck

.PHONY: gofmt
gofmt:
	GOFMT_RESULT=$$(gofmt -l $(GO_FILES)) && \
		[ -z "$$GOFMT_RESULT" ] || ( \
		echo "Error: the following files are not formatted: $$GOFMT_RESULT" && \
		false \
	)

.PHONY: golint
golint: $(GOLINT)
	$(GOLINT) ./...
$(GOLINT):
	go install golang.org/x/lint/golint@latest

.PHONY: staticcheck
staticcheck: $(STATICCHECK)
	$(STATICCHECK) ./...
$(STATICCHECK):
	go install honnef.co/go/tools/cmd/staticcheck@latest
