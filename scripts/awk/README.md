# AWK Scripts

This directory contains AWK scripts used by the Makefile system for generating help and target information.

## Files

### `help.awk`
- **Purpose**: Generates the formatted help output for `make help` command
- **Features**:
  - Parses all Makefile targets with `##` comments
  - Groups targets by categories (defined by `##@` headers)
  - Dynamically converts internal target names (e.g., `_install.*`) to user-friendly names based on filename
  - Supports dynamic prefix extraction from Makefile names

### `targets.awk`
- **Purpose**: Generates the formatted target list for `make targets` command
- **Features**:
  - Lists all documented targets from individual Makefiles
  - Applies the same dynamic prefix conversion as `help.awk`
  - Used in loops to process each Makefile separately

## Dynamic Prefix Feature

Both scripts support dynamic prefix generation based on Makefile names:

- `tools.mk` → `_install.grpc` becomes `tools.install.grpc`
- `service.mk` → targets remain unchanged
- `database.mk` → `_install.migrate` would become `database.install.migrate`

This allows for flexible naming without hardcoding specific file or category names.

## Usage

These scripts are automatically invoked by:
- `make help` - uses `help.awk`
- `make targets` - uses `targets.awk`

## Syntax Requirements

For targets to appear in help output, they must follow this format:
```makefile
target-name: ## Description of what this target does
```

For category headers:
```makefile
##@ Category Name
```
