# This script is called by `make help` to generate the help output.
# It parses all makefiles and prints a formatted list of targets.

BEGIN {
    # Match a colon, optional whitespace, and then ##.
    FS = ":[ \t]*##";
    current_category = "";
    # Print a static header.
    printf "\nUsage:\n  make \033[36m<TARGETS> <OPTIONS>\033[0m\n";
    printf "\n\033[35mTargets:\033[0m\n";
}

# Process category headers.
# A line starting with ##@ is a category header.
/^##@/ {
    category_name = substr($0, 5);
    printf "\n\033[1m%s\033[0m\n", category_name;
    current_category = tolower(category_name);
    next;
}

# Process any line that is a documented target.
# It must not start with whitespace or a hash, and must contain `: ##`.
/^[^\t #].*:[ \t]*##/ {
    target = $1;
    comment = $2;

    # Check if it's a variable-based target, e.g., `$(VAR)`
    if (target ~ /^\$\(/) {
        var = substr(target, 3, length(target)-3);
        # If the category is "tools", format as `tools.install.ci`.
        # Otherwise, format as `some-other-thing`.
        if (current_category == "tools") {
            gsub("_", ".", var);
        } else {
            gsub("_", "-", var);
        }
        printf "  \033[36m%-45s\033[0m%s\n", tolower(var), comment;
    } 
    # Else, it's a simple target, e.g., `build` or `target: prerequisite`
    else {
        # If there are prerequisites, only print the target name itself.
        split(target, arr, ":");
        printf "  \033[36m%-45s\033[0m%s\n", arr[1], comment;
    }
}

# After processing all files, print the footer.
END {
    # The USAGE_OPTIONS variable is passed from the Makefile environment.
    if (ENVIRON["USAGE_OPTIONS"]) {
        printf "%s\n", ENVIRON["USAGE_OPTIONS"];
    }
}
