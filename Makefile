.PHONY: create_files
create_files:
	./createFiles.sh ./hot 10 10

.PHONY: clean_files
clean_files:
	rm -rf ./hot/* ./backup/*

.PHONY: compare_files
compare_files:
	./compareFiles.sh ./hot ./backup

.PHONY: build
build:
	go build -o main cmd/*
