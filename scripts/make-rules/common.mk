
# include the common make file
ifeq ($(origin PROJ_ROOT_DIR),undefined)
PROJ_ROOT_DIR :=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
endif

include $(PROJ_ROOT_DIR)/scripts/make-rules/common-versions.mk


# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /usr/bin/env bash -o errexit -o pipefail +o nounset
.SHELLFLAGS = -ec

# It's necessary to set the errexit flags for the bash shell.
export SHELLOPTS := errexit


# ==============================================================================
# Build options
#
PRJ_SRC_PATH :=github.com/costa92/go-protoc/v2

COMMA := ,
SPACE :=
SPACE +=



ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR := $(PROJ_ROOT_DIR)/_output
$(shell mkdir -p $(OUTPUT_DIR))
endif

ifeq ($(origin LOCALBIN),undefined)
LOCALBIN := $(OUTPUT_DIR)/bin
$(shell mkdir -p $(LOCALBIN))
endif

ifeq ($(origin TOOLS_DIR),undefined)
TOOLS_DIR := $(OUTPUT_DIR)/tools
$(shell mkdir -p $(TOOLS_DIR))
endif

ifeq ($(origin TMP_DIR),undefined)
TMP_DIR := $(OUTPUT_DIR)/tmp
$(shell mkdir -p $(TMP_DIR))
endif


# set the version number. you should not need to do this
# for the majority of scenarios.
ifeq ($(origin VERSION), undefined)
# Current version of the project.
  VERSION := $(shell git describe --tags --always --match='v*')
  ifneq (,$(shell git status --porcelain 2>/dev/null))
    VERSION := $(VERSION)-dirty
  endif
endif

# Minimum test coverage
ifeq ($(origin COVERAGE),undefined)
COVERAGE := 60
endif

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
GOPATH ?= $(shell go env GOPATH)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif


# Makefile settings
#
# We don't need make's built-in rules.
MAKEFLAGS += --no-builtin-rules
ifeq ($(V),1)
  $(warning ***** starting Makefile for goal(s) "$(MAKECMDGOALS)")
  $(warning ***** $(shell date))
else
  # If we're not debugging the Makefile, don't echo recipes.]
  MAKEFLAGS += -s --no-print-directory
endif

# Linux command settings
FIND := find . ! -path './third_party/*' ! -path './vendor/*'
XARGS := xargs --no-run-if-empty


# Helper function to get dependency version from go.mod
get_go_version = $(shell go list -m $1 | awk '{print $$2}')
define go_install
$(info ===========> Installing $(1)@$(2))
$(GO) install $(1)@$(2)
endef


# Copy githook scripts when execute makefile
COPY_GITHOOK:=$(shell cp -f githooks/* .git/hooks/)

# Specify components which need certificate
ifeq ($(origin CERTIFICATES),undefined)
CERTIFICATES=go-protoc-apiserver admin
endif

MANIFESTS_DIR=$(PROJ_ROOT_DIR)/manifests
SCRIPTS_DIR=$(PROJ_ROOT_DIR)/scripts


APIROOT ?= $(PROJ_ROOT_DIR)/pkg/api
APISROOT ?= $(PROJ_ROOT_DIR)/pkg/apis