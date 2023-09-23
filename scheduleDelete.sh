#!/bin/bash

# Check if the required arguments are provided
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <directory> <filename>"
    exit 1
fi

# Variables from arguments
DIR="$1"
FILENAME="$2"

# Get ISODATETIME for 5 seconds into the future
ISODATETIME=$(date -d "now + 5 seconds" +'%Y-%m-%dT%H:%M:%S%:z')

# Construct full filename
FULL_FILENAME="${DIR}/delete_${ISODATETIME}_${FILENAME}"

# Create the file
touch "${FULL_FILENAME}"

# Confirm
echo "File created: ${FULL_FILENAME}"
