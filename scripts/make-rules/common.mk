
# include the common make file
ifeq ($(origin PROJECT_ROOT),undefined)
PROJECT_ROOT :=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
endif

# include the common versions make file
include $(PROJECT_ROOT)/scripts/make-rules/common-versions.mk


# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /usr/bin/env bash -o errexit -o pipefail +o nounset
.SHELLFLAGS = -ec

# It's necessary to set the errexit flags for the bash shell.
export SHELLOPTS := errexit


# ==============================================================================
# Build options
#

COMMA := ,
SPACE :=
SPACE +=

ifeq ($(origin OUTPUT_DIR),undefined)
OUTPUT_DIR := $(PROJECT_ROOT)/_output
$(shell mkdir -p $(OUTPUT_DIR))
endif

ifeq ($(origin LOCALBIN),undefined)
LOCALBIN := $(OUTPUT_DIR)/bin
$(shell mkdir -p $(LOCALBIN))
endif
