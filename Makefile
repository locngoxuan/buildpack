GOCMD=go
BINARY_NAME=bpp
VERSION?=2.1.1
PWD=$(shell pwd)

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: help clean build

default: help

clean:
	rm -fr ./bin

linux:
	docker run -v $(PWD):/workdir \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=1 \
		xuanloc0511/cgo_base:1.0.0 \
		go build -ldflags="-s -w -X main.version=${VERSION}" -o bin/linux/${BINARY_NAME} -a .

wins:
	docker run -v $(PWD):/workdir \
		-e GOOS=windows \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=1 \
		-e CROSS_TRIPLE=x86_64-w64-mingw32 \
		xuanloc0511/cgo_windows:1.0.0 \
		go build -ldflags="-s -w -X main.version=${VERSION}" -o bin/wins/${BINARY_NAME}.exe -a .

darwin:
	docker run -v $(PWD):/workdir \
		-e GOOS=darwin \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=1 \
		-e CROSS_TRIPLE=x86_64-apple-darwin \
		xuanloc0511/cgo_darwin:1.0.0 \
		go build -ldflags="-s -w -X main.version=${VERSION}" -o bin/darwin/${BINARY_NAME} -a .

docker:
	docker build -t xuanloc0511/buildpack_base:1.0.0 \
		-f ./dockers/Dockerfile .

help:
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@echo "  ${YELLOW}linux           ${RESET} ${GREEN}Build your project and put the output binary in bin/linux/$(BINARY_NAME)${RESET}"
	@echo "  ${YELLOW}wins            ${RESET} ${GREEN}Build your project and put the output binary in bin/wins/$(BINARY_NAME)${RESET}"
	@echo "  ${YELLOW}darwin          ${RESET} ${GREEN}Build your project and put the output binary in bin/darwin/$(BINARY_NAME)${RESET}"
	@echo "  ${YELLOW}clean           ${RESET} ${GREEN}Remove build related file${RESET}"
	@echo "  ${YELLOW}help            ${RESET} ${GREEN}Show this help message${RESET}"