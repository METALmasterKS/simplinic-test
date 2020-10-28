# Package dir
GO_DIR ?= $(CURDIR)

# Package import path
GO_PKG ?= $(shell go list -e -m -f "{{ .Dir }}" 2> /dev/null || go list -e -f "{{ .ImportPath }}" 2> /dev/null || pwd)

# Package version
GO_VER ?= $(shell date -u +%Y-%m-%d.%H:%M:%S)

# Package binary prefix
GO_BIN ?= $(shell basename $$(dirname $(GO_PKG)))

# Package binnary delimiter
GO_DEL ?= -

# .env
-include .env

all: help

.PHONY: help # Show list of targets with description. You're looking at it
help:
	@grep "^.PHONY: .* #" Makefile | sed "s/\.PHONY: \(.*\) # \(.*\)/\1 - \2/" | expand -t20

.PHONY: build # Compile application source code to binaries files
build:
	@echo "Build binaries"
		go build -gcflags="-trimpath=$(GO_DIR)" \
				 -asmflags="-trimpath=$(GO_DIR)" \
				 -ldflags "-X main.Version=$(GO_VER)" \
				 -o "$(GO_DIR)/.bin/" "$${CMD}" \

.PHONY: generate # Generate auto generated code
generate:
	@echo "Run easyjson"
	@go generate easyjson ./...

.PHONY: install # Install dependencies
install: conf-gomod
	@echo "Install dependencies"
	go mod download

.PHONY: update # Update dependencies
update: conf-gomod
	@echo "Update dependencies"
	go get -u -t all
	@echo "Optimize dependencies"
	go mod tidy -v

.PHONY: test # Run application tests
test:
	@echo "Run unit tests"
	@go test -v ./...


.PHONY: clean # Delete auto generated files
clean:
	@echo "Cleanup"
	@find . -type f -name "*easyjson*" -delete

.PHONY: docker-build # Docker build
docker-build: generate build
	@echo "Docker"
	@cp .bin/* .docker/files/project/
	@docker build .docker/ -t simplinic

