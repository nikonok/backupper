#!/bin/bash

# Check the number of arguments
if [[ $# -ne 3 ]]; then
    echo "Usage: $0 <directory> <number_of_files> <file_size_in_mb>"
    exit 1
fi

# Get the arguments
dir=$1
num_files=$2
file_size=$3

# Check if the specified directory exists
if [[ ! -d $dir ]]; then
    echo "Error: Directory $dir does not exist!"
    exit 1
fi

# Create the specified number of files with the given size
for (( i=1; i<=num_files; i++ )); do
    file_path="$dir/file_$i"
    dd if=/dev/zero of=$file_path bs=1M count=$file_size 2>/dev/null
    echo "Created $file_path with size ${file_size}MB"
done

echo "Done!"
