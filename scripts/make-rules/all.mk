# This file includes all other makefiles to centralize rule management.

# Include project-specific commands.
include scripts/make-rules/project.mk

# Include core build tools and utility commands.
include scripts/make-rules/tools.mk

# Include Go-specific build and formatting commands.
include scripts/make-rules/golang.mk

# Include documentation commands.
include scripts/make-rules/docs.mk

# Conditionally include service makefile if it exists.
# This allows for optional, user-defined service management commands.
ifneq ($(wildcard scripts/make-rules/service.mk),)
include scripts/make-rules/service.mk
endif