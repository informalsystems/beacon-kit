#!/usr/bin/make -f
ifeq ($(VERSION),)
  VERSION := $(shell git describe --tags --always --match "v*")
endif

COMMIT = $(shell git log -1 --format='%H')
CURRENT_DIR = $(shell pwd)
OUT_DIR ?= $(CURDIR)/build/bin
TESTNAME = beacon
TESTAPP = beacond
TESTAPP_FILES_DIR = testing/files
TESTAPP_CMD_DIR = cmd/$(TESTAPP)
PROJECT_NAME = $(shell git remote get-url origin | xargs basename -s .git)

# process build tags
build_tags = netgo

ifeq (legacy,$(findstring legacy,$(COSMOS_BUILD_OPTIONS)))
  build_tags += app_v1
endif

# DB backend selection
ifeq (cleveldb,$(findstring cleveldb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += gcc
endif
ifeq (badgerdb,$(findstring badgerdb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += badgerdb
endif
# handle rocksdb
ifeq (rocksdb,$(findstring rocksdb,$(COSMOS_BUILD_OPTIONS)))
  CGO_ENABLED=1
  build_tags += rocksdb grocksdb_clean_link
endif
# handle boltdb
ifeq (boltdb,$(findstring boltdb,$(COSMOS_BUILD_OPTIONS)))
  build_tags += boltdb
endif

# always include pebble
build_tags += pebbledb

# always include blst
build_tags += blst
build_tags += bls12381

# always include ckzg
build_tags += ckzg
build_tags += cgo

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=$(TESTNAME) \
		-X github.com/cosmos/cosmos-sdk/version.AppName=$(TESTAPP) \
		-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		-X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

build_tags += $(BUILD_TAGS)

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

# Check for debug option
ifeq (debug,$(findstring debug,$(COSMOS_BUILD_OPTIONS)))
  BUILD_FLAGS += -gcflags "all=-N -l"
endif

# This allows us to reuse the build target steps for both go build and go install
BUILD_TARGETS := build install

## Build:
build: BUILD_ARGS=-o $(OUT_DIR)/$(TESTAPP) ## build `beacond`

$(BUILD_TARGETS): $(OUT_DIR)/
	@echo "Building ${TESTAPP_CMD_DIR}"
	@go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) $(TESTAPP_CMD_DIR)/*.go

$(OUT_DIR)/:
	mkdir -p $(OUT_DIR)/

# Variables
ARCH ?= $(shell uname -m)
ifeq ($(ARCH),)
	ARCH = arm64
endif
IMAGE_NAME ?= $(TESTAPP)

# Docker Paths
DOCKERFILE = ./Dockerfile
DOCKERFILE_E2E = ./Dockerfile.e2e

build-docker: ## build a docker image containing `beacond`
	@echo "Build a release docker image for the Cosmos SDK chain..."
	docker build \
	--platform linux/$(ARCH) \
	--build-arg GIT_COMMIT=$(shell git rev-parse HEAD) \
	--build-arg GIT_VERSION=$(VERSION) \
	--build-arg GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD) \
	--build-arg GOOS=linux \
	--build-arg GOARCH=$(ARCH) \
	-f ${DOCKERFILE} \
	-t $(IMAGE_NAME):$(VERSION) \
	.

push-docker-github: ## push the docker image to the ghcr registry
	@echo "Push the release docker image to the ghcr registry..."
	docker tag $(IMAGE_NAME):$(VERSION) ghcr.io/berachain/beacon-kit:$(VERSION)
	docker push ghcr.io/berachain/beacon-kit:$(VERSION)

build-docker-e2e: ## build a docker image containing `beacond` used in the e2e tests
	@echo "Build an e2e docker image..."
	docker build \
	--build-arg IMAGE_NAME=$(IMAGE_NAME) \
	--build-arg VERSION=$(VERSION) \
	-f ${DOCKERFILE_E2E} \
	-t cometbft/e2e-node:$(VERSION) \
	./testing # work around .dockerignore restrictions in the root folder

build-generator: ## build e2e generator
	@echo "Build the e2e generator..."
	@echo "WARNING: go.mod and go.sum will be updated. Revert the changes if you do not want the generator requirements included."
	@go get github.com/cometbft/cometbft/test/e2e/generator
	@go build -mod=readonly -o $(OUT_DIR)/generator github.com/cometbft/cometbft/test/e2e/generator

# Hack: force the same viper version as CometBFT.
# The berachain cometbft fork is on an incompatible viper version.
# TODO: Update viper in the cometbft fork.
COMETBFT_DEPENDENCY_PATH := $(shell go list -m -f '{{ .Dir }}' github.com/cometbft/cometbft)
COMETBFT_VIPER_VERSION := "$(shell go list -m -f '{{ .Version }}' -modfile "$(COMETBFT_DEPENDENCY_PATH)/go.mod" github.com/spf13/viper)"
BEACONKIT_VIPER_VERSION := "$(shell go list -m -f '{{ .Version }}' github.com/spf13/viper)"

# Obsolete, remove!
build-runner-comet: ## build e2e runner
	@echo "Build the e2e runner..."

# Hack: Force the same viper version.
	@echo "CometBFT dependency path: $(COMETBFT_DEPENDENCY_PATH)"
	@test -d "$(COMETBFT_DEPENDENCY_PATH)" || go mod download
	@echo "Beaconkit Viper version: $(BEACONKIT_VIPER_VERSION)"
	@echo "CometBFT Viper version: $(COMETBFT_VIPER_VERSION)"
	@go get "github.com/spf13/viper@$(COMETBFT_VIPER_VERSION)"
# End of hack.
	@go build -mod=readonly -o $(OUT_DIR)/runner github.com/cometbft/cometbft/test/e2e/runner
# Hack: restore original viper version.
	@go get "github.com/spf13/viper@$(BEACONKIT_VIPER_VERSION)"
# End of hack.

build-runner: ## build e2e runner
	@echo "Build the e2e runner..."
	@go build -mod=readonly -o $(OUT_DIR)/runner ./testing/runner

build-e2e:
	@$(MAKE) build-docker VERSION=local-version \
    		build-docker-e2e VERSION=local-version \
    		build-generator \
    		build-runner
