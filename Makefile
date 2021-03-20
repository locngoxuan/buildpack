GOCMD=go
BINARY_NAME=bpp
VERSION?=2.1.1
PWD=$(shell pwd)
BASE_IMAGE=xuanloc0511/buildpack_base:$(VERSION)

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: help clean build

default: help

clean:
	rm -fr ./bin

linux:
	docker run -it --rm -v $(PWD):/buildpack \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=1 \
		--workdir="/buildpack" \
		$(BASE_IMAGE) \
		go build -ldflags="-s -w -X main.version=${VERSION}" -o bin/linux/${BINARY_NAME} ./cmd/.

wins:
	docker run -it --rm -v $(PWD):/buildpack \
		-e GOOS=windows \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=1 \
		--workdir="/buildpack" \
		$(BASE_IMAGE) \
		go build -ldflags="-s -w -X main.version=${VERSION}" -o bin/wins/${BINARY_NAME}.exe ./cmd/.

darwin:
	docker run -it --rm -v $(PWD):/buildpack \
		-e GOOS=darwin \
		-e GOARCH=amd64 \
		-e CGO_ENABLED=1 \
		--workdir="/buildpack" \
		$(BASE_IMAGE) \
		go build -ldflags="-s -w -X main.version=${VERSION}" -o bin/darwin/${BINARY_NAME} ./cmd/.

docker:
	docker build -t $(BASE_IMAGE) \
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