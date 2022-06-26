# (C) Copyright 2017 Mario Trangoni
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

GO                      ?= GO111MODULE=on go
GOPATH                  := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
PROMU                   ?= $(GOPATH)/bin/promu
GOLINTER                ?= $(GOPATH)/bin/golangci-lint
GO_VERSION              ?= 1.18
pkgs                    = $(shell $(GO) list ./... | grep -v /vendor/)
TARGET                  ?= flexlm_exporter
DOCKER_IMAGE_NAME       ?= mjtrangoni/flexlm_exporter
DOCKER_IMAGE_TAG        ?= $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))

PREFIX                  ?= $(shell pwd)
BIN_DIR                 ?= $(shell pwd)

.PHONY: all
all: clean format vet golangci build test

.PHONY: test
test:
	@echo ">> running tests"
	@$(GO) test -v $(pkgs)

.PHONY: format
format:
	@echo ">> formatting code"
	@$(GO) fmt $(pkgs)

.PHONY: vet
vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

.PHONY: golangci $(GOLINTER)
golangci: $(GOLINTER)
	@echo ">> linting code"
	@$(GOLINTER) run --config ./.golangci.yml

.PHONY: build
build: $(PROMU)
	@echo ">> building binaries"
	@$(PROMU) build --prefix $(PREFIX)

.PHONY: clean
clean:
	@echo ">> Cleaning up"
	@find . -type f -name '*~' -exec rm -fv {} \;
	@$(RM) $(TARGET)
	@$(RM) $(TARGET).exe
	@$(RM) -rv ./.build

.PHONY: docker
docker:
	@echo ">> building docker image"
	@docker build -t "$(DOCKER_IMAGE_NAME)" -t "$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" .

.PHONY: promu
$(GOPATH)/bin/promu promu:
	@GOOS=$(shell uname -s | tr A-Z a-z) \
		GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
		$(GO) install github.com/prometheus/promu@v0.13.0

.PHONY: tarball
tarball: promu
	@echo ">> building release tarball"
	$(PROMU) tarball --prefix $(PREFIX) $(BIN_DIR)

.PHONY: crossbuild
crossbuild: promu
	@echo ">> crossbuilding"
	$(PROMU) crossbuild --go=$(GO_VERSION) $(BIN_DIR)

.PHONY: golangci-lint lint
$(GOPATH)/bin/golangci-lint lint:
	@GOOS=$(shell uname -s | tr A-Z a-z) \
		GOARCH=$(subst x86_64,amd64,$(patsubst i%86,386,$(shell uname -m))) \
		$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.46.2
