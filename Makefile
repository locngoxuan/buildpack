.PHONY: vendor get format test clean build

BUILD_NAME=buildpack
INSTALL_DIR=/usr/local/bin

build:
	env CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${BUILD_NAME} -a -v .

linux:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${BUILD_NAME}-linux -a -v .

windows:
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/${BUILD_NAME}-wins.exe -a -v .

macos:
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${BUILD_NAME}-darwin -a -v .

compress:
	upx --brute ./bin/${BUILD_NAME}

compress-linux:
	upx --brute ./bin/${BUILD_NAME}-linux

compress-darwin:
	upx --brute ./bin/${BUILD_NAME}-darwin

compress-all: compress compress-linux compress-darwin

install:
	cp -r ./bin/${BUILD_NAME} ${INSTALL_DIR}/${BUILD_NAME}

test: get
	go test -v .

vendor:
	go mod vendor

get:
	go get -t -v ./...

format:
	find . -name \*.go -type f -exec gofmt -w {} \;

clean:
	rm -rf bin

all: clean build linux windows macos compress-all

one: clean build compress