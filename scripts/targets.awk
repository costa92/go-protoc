# This script is called by `make targets` to generate the targets list for a single makefile.
# It's called inside a loop for each makefile.

# Skip category headers, as the Makefile loop handles printing the filename.
/^##@/ {next}

# Process simple targets like `build: ## ...`
/^[a-zA-Z0-9._-]+:.*?##/ {
    printf "  \033[36m%-45s\033[0m %s\n", $1, $2
}

# Process variable-based targets like `$(VAR): ## ...`
/^\$\([a-zA-Z0-9._-]+\):.*?##/ {
    var=substr($1, 3, length($1)-3)
    # The category is passed in from the Makefile loop.
    if (category == "tools") {
        gsub("_", ".", var)
    } else {
        gsub("_", "-", var)
    }
    printf "  \033[36m%-45s\033[0m %s\n", tolower(var), $2
}

