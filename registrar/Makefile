# go option
GO             ?= go
GOIMPORTS      ?= goimports
JSONNETFMT     ?= jsonnetfmt
PKG            := $(GO) mod vendor
APP_VERSION    := 1.0.0-$(shell git rev-parse HEAD)
LDFLAGS        := -w -s
GOFLAGS        :=
GOPRIVATE      := 
GOPROXY        := https://proxy.golang.org
GO_EXTRA_FLAGS := -v -tags=or_dev
TAGS           :=
BINDIR         := $(CURDIR)/bin
BIN_NAME       := authz
PKGDIR         := github.com/jaredallard-home/worker-nodes/registrar
CGO_ENABLED    := 1
TOOL_DEPS      := ${GO} ${GOIMPORTS} ${JSONNETFMT}
BENCH_FLAGS    := -benchtime=10x

.PHONY: default
default: build

.PHONY: check_deps
check_deps:
	@true

.PHONY: pre-commit
pre-commit: fmt

.PHONY: build
build: gogenerate gobuild

.PHONY: test
test:
	GOPROXY=$(GOPROXY) GOPRIVATE=$(GOPRIVATE) go test ./...

.PHONY: docs
docs:
	@echo "Not done yet"

.PHONY: dep
dep:
	@echo " ===> Installing dependencies via <=== "
	GOPROXY=$(GOPROXY) GOPRIVATE=$(GOPRIVATE) $(PKG)

.PHONY: gogenerate
gogenerate: check_deps
	GOPROXY=$(GOPROXY) GOPRIVATE=$(GOPRIVATE) $(GO) generate ./...

.PHONY: gobuild
gobuild: check_deps
	@echo " ===> building releases in ./bin/... <=== "
	mkdir -p $(BINDIR)
	GOPROXY=$(GOPROXY) GOPRIVATE=$(GOPRIVATE) CGO_ENABLED=$(CGO_ENABLED) $(GO) build -o $(BINDIR)/ -ldflags "$(LDFLAGS)" $(GO_EXTRA_FLAGS) $(PKGDIR)/...
.PHONY: docker-build
docker-build:
	DOCKER_BUILDKIT=1 docker build --ssh default -t jaredallard/registrar -f Dockerfile . --build-arg VERSION=${APP_VERSION}

.PHONY: fmt
fmt:
	@echo " ==> Running goimports <=="
	find  . -path ./vendor -prune -o -type f -name '*.go' -print | xargs -n 1 ${GOIMPORTS} -w
	@echo " ==> Running jsonnetfmt <=="
	find . -name '*.jsonnet' | xargs -n 1 jsonnetfmt -i
