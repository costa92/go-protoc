# This script is called by `make help` to generate the help output.
# It parses all makefiles and prints a formatted list of targets.

@include "scripts/awk/common.awk"

BEGIN {
    FS = ":[ \t]*##";  # Split on colon + spaces/tabs + ##
    current_category = "";

    printf "\nUsage:\n  make " COLOR_CYAN "<TARGETS> <OPTIONS>" COLOR_RESET "\n";
    printf "\n" COLOR_MAGENTA "Targets:" COLOR_RESET "\n";
}

# Process category headers
/^##@/ {
    current_category = process_category_header($0, "\n");
    next;
}

# Process documented targets
/^[^	 #].*:[ 	]*##/ {
    target = $1;
    comment = $2;
    file_prefix = get_file_prefix(FILENAME);

    # Variable form: $(VAR)
    var = extract_var_target(target);
    if (var != "") {
        var = format_target(var, file_prefix);
        format_target_line(tolower(var), comment, "  ");
    }
    # Simple target
    else {
        target_name = extract_simple_target(target);
        target_name = format_target(target_name, file_prefix);
        format_target_line(target_name, comment, "  ");
    }
}

# Print footer
END {
    if (ENVIRON["USAGE_OPTIONS"]) {
        printf "%s\n", ENVIRON["USAGE_OPTIONS"];
    }
}
