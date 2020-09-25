CMD = agent server
TARGET = build
PROJECTNAME = Talos
PACKAGES ?= $(shell go list ./... | grep -v /vendor/)
GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")
GOFMT ?= gofmt "-s"
DATE = `date +%FT%T%z`
VERSION = 0.1

MAKEFILES += --silent

all: help

## fmt: format go files
fmt:
	$(GOFMT) -w $(GOFILES)

fmt-check:
	@diff=$$($(GOFMT) -d $(GOFILES)); \
	if [ -n "$$diff" ]; then \
		echo "Please run 'make fmt' and commit the result:"; \
		echo "$${diff}"; \
		exit 1; \
	fi;

$(CMD): mod clean
	@echo "  >  Building $(@) binary..."
	@go build -o build/$@ ./cmd/$@/$@.go

## mod: go mod vendor
mod:
	@echo "  >  Go mod vendor..."
	@go mod vendor

## build: build server/agent
build: $(CMD) $(TARGET)

## build-linux: build linux binary
build-linux: mod clean
	@echo "  >  Building agent linux binary..."
	@GOOS=linux go build -o build/agent ./cmd/agent/agent.go
	@echo "  >  Building server linux binary..."
	@GOOS=linux go build -o build/server ./cmd/server/server.go

## run-server: run server -cd .
run-server:
	@echo "  >  Run server with conf path ./ "
	@go run ./cmd/server/server.go run -cd .

## run-agent: run agent -cd .
run-agent:
	@echo "  >  Run agent with conf path ./ "
	@go run ./cmd/agent/agent.go run -cd .

## clean: clean build cache
clean:
	@echo "  >  Cleaning build cache..."
	@rm -rf ./build/*

.PHONY: all clean fmt-check help

help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@sed -n 's/^##/ >/p' $< | column -t -s ':' | sed -e 's/^/ /'
	@echo
