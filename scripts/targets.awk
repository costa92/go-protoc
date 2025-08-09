# This script is called by `make targets` to generate the targets list for a single makefile.
# It's called inside a loop for each makefile.

# Skip category headers, as the Makefile loop handles printing the filename.
/^##@/ {next}

# Process simple targets like `build: ## ...`
/^[a-zA-Z0-9._-]+:.*?##/ {
    target_name = $1;
    # Extract file prefix from filename (e.g., tools.mk -> tools)
    file_prefix = "";
    if (FILENAME ~ /\.mk$/) {
        split(FILENAME, path_parts, "/");
        filename = path_parts[length(path_parts)];
        sub(/\.mk$/, "", filename);
        file_prefix = filename;
    }

    # If we have a file prefix, convert '_install.*' to '{prefix}.install.*'
    # and '_verify.*' to '{prefix}.verify.*' for display.
    if (file_prefix != "") {
        if (target_name ~ /^_install\./) {
            sub(/^_/, file_prefix ".", target_name);
        }
        else if (target_name ~ /^_verify\./) {
            sub(/^_/, file_prefix ".", target_name);
        }
    }
    printf "  \033[36m%-45s\033[0m %s\n", target_name, $2
}

# Process variable-based targets like `$(VAR): ## ...`
/^\$\([a-zA-Z0-9._-]+\):.*?##/ {
    var=substr($1, 3, length($1)-3)
    # Extract file prefix from filename (e.g., tools.mk -> tools)
    file_prefix = "";
    if (FILENAME ~ /\.mk$/) {
        split(FILENAME, path_parts, "/");
        filename = path_parts[length(path_parts)];
        sub(/\.mk$/, "", filename);
        file_prefix = filename;
    }

    # If we have a file prefix, use it for transformation
    if (file_prefix != "") {
        # Convert _install.* to {prefix}.install.*
        if (var ~ /^_install\./) {
            sub(/^_/, file_prefix ".", var);
        }
        # Convert _verify.* to {prefix}.verify.*
        else if (var ~ /^_verify\./) {
            sub(/^_/, file_prefix ".", var);
        }
        # For other patterns, keep dots
        gsub("_", ".", var);
    } else {
        gsub("_", "-", var);
    }
    printf "  \033[36m%-45s\033[0m %s\n", tolower(var), $2
}

