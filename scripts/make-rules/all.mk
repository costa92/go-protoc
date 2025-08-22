# This file includes all other makefiles to centralize rule management.

# Include project-specific commands.
include scripts/make-rules/project.mk

# Include build system commands.
include scripts/make-rules/build.mk

# Include core build tools and utility commands.
include scripts/make-rules/tools.mk

@echo "==> DEBUG: Including deploy.mk"
include scripts/make-rules/deploy.mk

# Include Go-specific build and formatting commands.
include scripts/make-rules/golang.mk

# Include documentation commands.
include scripts/make-rules/docs.mk

# Include database management commands.
include scripts/make-rules/database.mk

# Conditionally include service makefile if it exists.
# This allows for optional, user-defined service management commands.
ifneq ($(wildcard scripts/make-rules/service.mk),)
include scripts/make-rules/service.mk
endif