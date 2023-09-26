SCRIPTS_DIR=scripts

.PHONY:all
all: install_deps build

.PHONY: create_files
create_files:
	${SCRIPTS_DIR}/createFiles.sh ./hot 10 10

.PHONY: clean_files
clean_files:
	rm -rf ./hot/* ./backup/* log.txt

.PHONY: compare_files
compare_files:
	${SCRIPTS_DIR}/compareFiles.sh ./hot ./backup

.PHONY: create_scheduled
create_scheduled:
	${SCRIPTS_DIR}/scheduleDelete.sh hot file_1

.PHONY: install_deps
install_deps:
	go mod download

.PHONY: build
build:
	go build -o main cmd/*

.PHONY: docker
docker:
	docker build -t backupper .

.PHONY: docker-run
docker-run:
	docker run -v $$(pwd):/app backupper

.PHONY: test
test:
	go test ./...
	${SCRIPTS_DIR}/test.sh
