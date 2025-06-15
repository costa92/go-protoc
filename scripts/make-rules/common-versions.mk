# ==============================================================================
# Versions used by all Makefiles
#

# 安装 wire 工具
WIRE_VERSION ?= $(call get_go_version,github.com/google/wire)

# 安装 kratos 工具
GO_APIDIFF_VERSION ?= v0.8.2
PROTOC_GEN_VALIDATE_VERSION ?= $(call get_go_version,github.com/envoyproxy/protoc-gen-validate)
PROTOC_GEN_OPENAPI_VERSION ?= v0.7.0
KRATOS_VERSION ?= $(call get_go_version,github.com/go-kratos/kratos/v2)

# 安装 logcheck 工具
LOGCHECK_VERSION ?= v0.8.1

# 安装 grpcurl 工具
GRPCURL_VERSION ?= v1.8.9

# 安装 code-generator 工具
CODE_GENERATOR_VERSION ?= $(call get_go_version,k8s.io/code-generator)