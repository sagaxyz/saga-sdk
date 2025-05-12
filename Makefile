# ──────────────────────────────────────────────────────────────────────────────
# Build configuration
# ──────────────────────────────────────────────────────────────────────────────

GIT_TAG        		:= $(shell git describe --tags --abbrev=0 --match "v[0-9]*" 2>/dev/null)
BUILD_METADATA 		:= $(shell echo $(GIT_TAG) | sed 's/^v//')
COMMIT_SHORT   		:= $(shell git rev-parse --short=7 HEAD)
export VERSION 		:= $(BUILD_METADATA)+$(COMMIT_SHORT)
export CMTVERSION 	:= $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::')
export COMMIT     	:= $(shell git log -1 --format='%H')
LEDGER_ENABLED 		?= true

BUILDDIR     ?= build
SIMAPP        = simapp
BINARY_NAME   = simd

BUILD_FLAGS   = -tags "netgo ledger app_v1" \
				-mod=readonly \
                -ldflags "\
                  -X github.com/cosmos/cosmos-sdk/version.Name=$(BINARY_NAME) \
                  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
                  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
                  -X github.com/cometbft/cometbft/version.TMCoreSemVer=$(CMTVERSION)" \
                -trimpath

GOBIN   := $(shell go env GOBIN)
GOPATH  := $(shell go env GOPATH)
DESTDIR := $(if $(GOBIN),$(GOBIN),$(GOPATH)/bin)


# ──────────────────────────────────────────────────────────────────────────────
# High‑level targets
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: all proto build install test clean
all: proto build

# generate protobufs
proto: proto-gen

# build the simd binary
build: build-dir
	@cd $(SIMAPP) && \
	go build -mod=readonly $(BUILD_FLAGS) \
	         -o ../$(BUILDDIR)/$(BINARY_NAME) ./simd

# install into your Go bin dir
install: build
	@echo "Installing $(BINARY_NAME) to $(DESTDIR)"
	install -m 0755 $(BUILDDIR)/$(BINARY_NAME) $(DESTDIR)/$(BINARY_NAME)

# run tests
test:
	go test -race -cover -coverprofile cp.out -count=1 -timeout=30s ./...

# wipe out build artifacts
clean:
	rm -rf $(BUILDDIR)


# ──────────────────────────────────────────────────────────────────────────────
# internal helpers
# ──────────────────────────────────────────────────────────────────────────────

# ensure build directory exists
.PHONY: build-dir
build-dir:
	mkdir -p $(BUILDDIR)


# protobuf generation
proto_ver        = 0.11.6
proto_image_name = ghcr.io/cosmos/proto-builder:$(proto_ver)
proto_image      = docker run --rm -v $(CURDIR):/workspace --workdir /workspace $(proto_image_name)

.PHONY: proto-gen
proto-gen:
	@echo "Generating Protobuf files"
	@$(proto_image) sh ./scripts/protocgen.sh
