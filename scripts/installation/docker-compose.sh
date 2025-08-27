#!/bin/bash

# ==============================================================================
# Docker Compose Installation Script
# This script installs Docker Compose on macOS and Linux systems
# ==============================================================================

set -e


# Default version (can be overridden by environment variable)
# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# 使用统一版本管理的 Docker Compose 版本
DOCKER_COMPOSE_VERSION="${DOCKER_COMPOSE_VERSION}"

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if Docker Compose is already installed
check_existing_installation() {
    if command -v docker-compose >/dev/null 2>&1; then
        local current_version=$(docker-compose --version 2>/dev/null | grep -oE 'v?[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        print_info "Docker Compose is already installed (version: ${current_version})"

        # Check if we should update
        if [[ "${current_version}" != "${DOCKER_COMPOSE_VERSION#v}" ]]; then
            print_warning "Current version (${current_version}) differs from target version (${DOCKER_COMPOSE_VERSION#v})"
            read -p "Do you want to update? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                print_info "Keeping current installation"
                exit 0
            fi
        else
            print_info "Target version is already installed"
            exit 0
        fi
    fi
}

# Function to install Docker Compose on macOS
install_macos() {
    print_info "Installing Docker Compose on macOS..."

    if command -v brew >/dev/null 2>&1; then
        print_info "Using Homebrew to install Docker Compose..."
        brew install docker-compose
    else
        print_error "Homebrew is required to install Docker Compose on macOS"
        print_info "Please install Homebrew first: https://brew.sh/"
        exit 1
    fi
}

# Function to install Docker Compose on Linux
install_linux() {
    print_info "Installing Docker Compose on Linux..."

    local arch=$(uname -m)
    local os=$(uname -s)
    local download_url="https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-${os}-${arch}"
    local temp_file="/tmp/docker-compose"
    local install_path="/usr/local/bin/docker-compose"

    print_info "Downloading Docker Compose ${DOCKER_COMPOSE_VERSION} for ${os}-${arch}..."

    if curl -L "${download_url}" -o "${temp_file}"; then
        print_info "Download completed successfully"

        # Make it executable
        chmod +x "${temp_file}"

        # Move to installation directory (requires sudo)
        if [[ $EUID -ne 0 ]]; then
            print_info "Installing to ${install_path} (requires sudo)..."
            sudo mv "${temp_file}" "${install_path}"
        else
            mv "${temp_file}" "${install_path}"
        fi

        print_info "Docker Compose installed to ${install_path}"
    else
        print_error "Failed to download Docker Compose"
        print_error "Download URL: ${download_url}"
        exit 1
    fi
}

# Function to verify installation
verify_installation() {
    print_info "Verifying Docker Compose installation..."

    if command -v docker-compose >/dev/null 2>&1; then
        local installed_version=$(docker-compose --version 2>/dev/null | grep -oE 'v?[0-9]+\.[0-9]+\.[0-9]+' | head -1)
        print_info "Docker Compose successfully installed!"
        print_info "Version: ${installed_version}"
        print_info "Location: $(which docker-compose)"
    else
        print_error "Docker Compose installation verification failed"
        exit 1
    fi
}

# Main installation function
main() {
    print_info "Starting Docker Compose installation..."
    print_info "Target version: ${DOCKER_COMPOSE_VERSION}"

    # Check if already installed
    check_existing_installation

    # Detect operating system and install accordingly
    case "$(uname)" in
        "Darwin")
            install_macos
            ;;
        "Linux")
            install_linux
            ;;
        *)
            print_error "Unsupported operating system: $(uname)"
            print_error "This script supports macOS (Darwin) and Linux only"
            exit 1
            ;;
    esac

    # Verify the installation
    verify_installation

    print_info "Docker Compose installation completed successfully!"
}

# Show help information
show_help() {
    cat << EOF
Docker Compose Installation Script

Usage: $0 [OPTIONS]

OPTIONS:
    -h, --help          Show this help message
    -v, --version       Show script version
    --docker-compose-version VERSION
                        Specify Docker Compose version to install (default: ${DOCKER_COMPOSE_VERSION})

ENVIRONMENT VARIABLES:
    DOCKER_COMPOSE_VERSION  Docker Compose version to install

EXAMPLES:
    $0                                          # Install default version
    $0 --docker-compose-version v2.28.1        # Install specific version
    DOCKER_COMPOSE_VERSION=v2.28.1 $0          # Install via environment variable

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--version)
            echo "Docker Compose Installation Script v1.0.0"
            exit 0
            ;;
        --docker-compose-version)
            DOCKER_COMPOSE_VERSION="$2"
            shift 2
            ;;
        *)
            print_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Run main function only when script is executed directly (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
