# This script is called by `make targets` to generate the targets list for a single makefile.
# It's called inside a loop for each makefile.

@include "scripts/awk/common.awk"

# Process category headers
/^##@/ {
    process_category_header($0, "\n  ");
    next;
}

# Process simple targets like `build: ## ...`
/^[a-zA-Z0-9._-]+:.*?##/ {
    target_name = extract_simple_target($1);
    file_prefix = get_file_prefix(FILENAME);
    target_name = format_target(target_name, file_prefix);
    format_target_line(target_name, $2, "  ");
}

# Process variable-based targets like `$(VAR): ## ...`
/^\$\([a-zA-Z0-9._-]+\):.*?##/ {
    var = extract_var_target($1);
    file_prefix = get_file_prefix(FILENAME);
    var = format_target(var, file_prefix);
    format_target_line(tolower(var), $2, "  ");
}
