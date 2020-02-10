.PHONY: vendor get format test clean build

SQLBUNDLE_BUILD=sqlbundle
BUILDPACK_BUILD=buildpack
INSTALL_DIR=/usr/local/bin

build:
	env CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${SQLBUNDLE_BUILD} -a ./sqlbundle/cmd
	env CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${BUILDPACK_BUILD} -a ./cmd

linux:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${SQLBUNDLE_BUILD}-linux -a ./sqlbundle/cmd
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${BUILDPACK_BUILD}-linux -a ./cmd

windows:
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/${SQLBUNDLE_BUILD}-wins.exe -a ./sqlbundle/cmd
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/${BUILDPACK_BUILD}-wins.exe -a ./cmd

macos:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${SQLBUNDLE_BUILD}-linux -a ./sqlbundle/cmd
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${BUILDPACK_BUILD}-darwin -a ./cmd

compress:
	upx --brute ./bin/${SQLBUNDLE_BUILD}
	upx --brute ./bin/${BUILDPACK_BUILD}

compress-linux:
	upx --brute ./bin/${SQLBUNDLE_BUILD}-linux
	upx --brute ./bin/${BUILDPACK_BUILD}-linux

compress-darwin:
	upx --brute ./bin/${SQLBUNDLE_BUILD}-darwin
	upx --brute ./bin/${BUILDPACK_BUILD}-darwin

compress-all: compress compress-linux compress-darwin

install:
	chmod 755 ./bin/${SQLBUNDLE_BUILD}
	chmod 755 ./bin/${BUILDPACK_BUILD}
	cp -r ./bin/${SQLBUNDLE_BUILD} ${INSTALL_DIR}/${SQLBUNDLE_BUILD}
	cp -r ./bin/${BUILDPACK_BUILD} ${INSTALL_DIR}/${BUILDPACK_BUILD}

test: get
	go test -v .

vendor:
	go mod vendor

get:
	go get -t -v ./...

format:
	find . -name \*.go -type f -exec gofmt -w {} \;

clean:
	rm -rf ./bin

all: clean build linux windows macos compress-all
