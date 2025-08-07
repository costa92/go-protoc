#!/bin/bash
# Generate development documentation from template

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

echo "===========> Generating development documentation..."

# Configuration
CURRENT_TIME=$(date '+%Y-%m-%d %H:%M:%S')
PROJECT_NAME="go-protoc"
MODULE_PATH="github.com/costa92/go-protoc/v2"
GITHUB_URL="https://github.com/costa92/go-protoc"

# Ensure docs directory exists
mkdir -p docs

# Template processing function
process_template() {
    local template_file="$1"
    local output_file="$2"
    
    # Use envsubst or simple sed for variables
    sed -e "s|{{.GenerateTime}}|$CURRENT_TIME|g" \
        "$template_file" > "$output_file"
}

# Process template
if [ -f "templates/dev-docs.template.md" ]; then
    process_template "templates/dev-docs.template.md" "docs/DEVELOPMENT.md"
    echo "Development documentation generated to docs/DEVELOPMENT.md"
else
    echo "ERROR: Template file templates/dev-docs.template.md not found" >&2
    exit 1
fi