# This script is called by `make help` to generate the help output.
# It parses all makefiles and prints a formatted list of targets.

BEGIN {
    FS = ":[ \t]*##";  # Split on colon + spaces/tabs + ##
    current_category = "";
    # List of special prefixes to replace
    special_prefixes = "_install\\.|_uninstall\\.|_verify\\.";

    printf "\nUsage:\n  make \033[36m<TARGETS> <OPTIONS>\033[0m\n";
    printf "\n\033[35mTargets:\033[0m\n";
}

# Process category headers
/^##@/ {
    category_name = substr($0, 5);
    printf "\n\033[1m%s\033[0m\n", category_name;
    current_category = tolower(category_name);
    next;
}

# Process documented targets
/^[^	 #].*:[ 	]*##/ {
    target = $1;
    comment = $2;

    # Get file prefix if in a .mk file
    file_prefix = "";
    if (FILENAME ~ /\.mk$/) {
        split(FILENAME, path_parts, "/");
        filename = path_parts[length(path_parts)];
        sub(/\.mk$/, "", filename);
        file_prefix = filename;
    }

    # Variable form: $(VAR)
    if (target ~ /^\$\(/) {
        var = substr(target, 3, length(target)-3);

        if (file_prefix != "") {
            # Replace special prefixes in one go
            if (var ~ "^(" special_prefixes ")") {
                sub(/^_/, file_prefix ".", var);
            }
            gsub("_", ".", var); # Convert underscores to dots
        } else {
            gsub("_", "-", var); # Convert underscores to hyphens
        }

        printf "  \033[36m%-45s\033[0m%s\n", tolower(var), comment;
    }
    # Simple target
    else {
        split(target, arr, ":");
        target_name = arr[1];

        if (file_prefix != "") {
            if (target_name ~ "^(" special_prefixes ")") {
                sub(/^_/, file_prefix ".", target_name);
            }
        }

        printf "  \033[36m%-45s\033[0m%s\n", target_name, comment;
    }
}

# Print footer
END {
    if (ENVIRON["USAGE_OPTIONS"]) {
        printf "%s\n", ENVIRON["USAGE_OPTIONS"];
    }
}
