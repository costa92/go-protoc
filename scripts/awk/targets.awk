# This script is called by `make targets` to generate the targets list for a single makefile.
# It's called inside a loop for each makefile.

# List of special prefixes to replace when file prefix is available
BEGIN {
    special_prefixes = "_install\\.|_uninstall\\.|_verify\\.";
}

# Function to apply file prefix rules and formatting
function format_target(name, file_prefix,   formatted) {
    if (file_prefix != "") {
        if (name ~ "^(" special_prefixes ")") {
            sub(/^_/, file_prefix ".", name);
        }
        gsub("_", ".", name); # Keep dots for other cases
    } else {
        gsub("_", "-", name); # No file prefix → underscores to hyphens
    }
    return name;
}

# Skip category headers, as the Makefile loop handles printing the filename
/^##@/ { next }

# Process simple targets like `build: ## ...`
/^[a-zA-Z0-9._-]+:.*?##/ {
    target_name = $1;
    file_prefix = "";
    if (FILENAME ~ /\.mk$/) {
        split(FILENAME, path_parts, "/");
        filename = path_parts[length(path_parts)];
        sub(/\.mk$/, "", filename);
        file_prefix = filename;
    }
    target_name = format_target(target_name, file_prefix);
    printf "  \033[36m%-45s\033[0m %s\n", target_name, $2;
}

# Process variable-based targets like `$(VAR): ## ...`
/^\$\([a-zA-Z0-9._-]+\):.*?##/ {
    var = substr($1, 3, length($1)-3);
    file_prefix = "";
    if (FILENAME ~ /\.mk$/) {
        split(FILENAME, path_parts, "/");
        filename = path_parts[length(path_parts)];
        sub(/\.mk$/, "", filename);
        file_prefix = filename;
    }
    var = format_target(var, file_prefix);
    printf "  \033[36m%-45s\033[0m %s\n", tolower(var), $2;
}
