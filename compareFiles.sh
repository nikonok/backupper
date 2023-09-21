#!/bin/bash

# Check the number of arguments
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <source_dir> <backup_dir>"
    exit 1
fi

source_dir="$1"
backup_dir="$2"

# Verify both arguments are directories
if [ ! -d "$source_dir" ] || [ ! -d "$backup_dir" ]; then
    echo "Both arguments should be directories."
    exit 1
fi

# Loop through each file in the source directory and compare it to the file in the backup directory
find "$source_dir" -type f -print0 | while IFS= read -r -d '' file; do
    rel_path="${file#$source_dir/}"
    backup_file="$backup_dir/$rel_path.bak"

    if [ ! -f "$backup_file" ]; then
        echo "Missing in backup: $rel_path"
        continue
    fi

    diff_output=$(diff "$file" "$backup_file")
    if [ $? -ne 0 ]; then
        echo "Difference found for: $rel_path"
        echo "$diff_output"
    fi
done

echo "Check complete."
