# ==============================================================================
# Makefile helper functions for golang
#
GO := go
# Minimum supported go version.
GO_MINIMUM_VERSION ?= 1.22


ifeq ($(PRJ_SRC_PATH),)
	$(error the variable PRJ_SRC_PATH must be set prior to including golang.mk)
endif
ifeq ($(PROJ_ROOT_DIR),)
	$(error the variable PROJ_ROOT_DIR must be set prior to including golang.mk)
endif

GIT_TREE_STATE:="dirty"
ifeq (, $(shell git status --porcelain 2>/dev/null))
    GIT_TREE_STATE="clean"
endif
GIT_COMMIT:=$(shell git rev-parse HEAD)