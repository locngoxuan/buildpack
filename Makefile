GOCMD=go
BINARY_NAME=bpp
VERSION?=2.1.1
OS=linux
ARCH=amd64

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: help clean build

default: help

clean:
	rm -fr ./bin

dev:
	mkdir -p bin
	go build  -ldflags="-X main.version=${VERSION}" -o bin/${BINARY_NAME} .

build:
	mkdir -p bin
	env GOOS=${BUILD_OS} GOARCH=${BUILD_ARCH} CGO_ENABLED=1 go build -ldflags="-s -w -X main.version=${VERSION}" -o bin/${BINARY_NAME} -a .

install:
	cp -r bin/${BINARY_NAME} /usr/bin/${BINARY_NAME}

help:
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@echo "  ${YELLOW}build           ${RESET} ${GREEN}Build your project and put the output binary in bin/$(BINARY_NAME)${RESET}"
	@echo "  ${YELLOW}clean           ${RESET} ${GREEN}Remove build related file${RESET}"
	@echo "  ${YELLOW}help            ${RESET} ${GREEN}Show this help message${RESET}"