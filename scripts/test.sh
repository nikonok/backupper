#!/bin/bash

cleanup() {
    kill $main_pid 2>/dev/null
}

trap cleanup EXIT

./main -log-level 3 &
main_pid=$!

make -s create_files
sleep 3

make -s compare_files

make -s create_scheduled
sleep 5

if [ -f "./backup/file_1.bak" ]; then
    echo "Error: file_1 exists in ./backup"
    exit 1
fi

make -s compare_files

touch hot/delete_file_2
sleep 3

if [ -f "./backup/file_2.bak" ]; then
    echo "Error: file_2 exists in ./backup"
    exit 1
fi

make -s compare_files
