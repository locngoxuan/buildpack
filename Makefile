.PHONY: sqlbundle

SQLBUNDLE_BUILD=sqlbundle
BUILDPACK_BUILD=buildpack
INSTALL_DIR=/usr/local/bin

#apply for develop
buildpack:
	env CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${BUILDPACK_BUILD} -a ./cmd

#apply for develop
sqlbundle:
	env CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${SQLBUNDLE_BUILD} -a ./sqlbundle/cmd

#apply on release
build_buildpack:
	env CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${BUILDPACK_BUILD} -a ./cmd
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${BUILDPACK_BUILD}-linux -a ./cmd
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/${BUILDPACK_BUILD}-wins.exe -a ./cmd
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${BUILDPACK_BUILD}-darwin -a ./cmd

#apply on release
build_sqlbundle:
	env CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${SQLBUNDLE_BUILD} -a ./sqlbundle/cmd
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${SQLBUNDLE_BUILD}-linux -a ./sqlbundle/cmd
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/${SQLBUNDLE_BUILD}-wins.exe -a ./sqlbundle/cmd
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${SQLBUNDLE_BUILD}-darwin -a ./sqlbundle/cmd

#apply on release
build_all: build_sqlbundle build_buildpack

#apply on release
compress_buildpack:
	upx --brute ./bin/${BUILDPACK_BUILD}
	upx --brute ./bin/${BUILDPACK_BUILD}-linux
	upx --brute ./bin/${BUILDPACK_BUILD}-darwin

#apply on release
compress_sqlbundle:
	upx --brute ./bin/${SQLBUNDLE_BUILD}
	upx --brute ./bin/${SQLBUNDLE_BUILD}-linux
	upx --brute ./bin/${SQLBUNDLE_BUILD}-darwin

#apply on release
compress-all: compress_sqlbundle compress_buildpack

install_sqlbundle:
	chmod 755 ./bin/${SQLBUNDLE_BUILD}
	cp -r ./bin/${SQLBUNDLE_BUILD} ${INSTALL_DIR}/${SQLBUNDLE_BUILD}

install_buildpack:
	chmod 755 ./bin/${BUILDPACK_BUILD}
	cp -r ./bin/${BUILDPACK_BUILD} ${INSTALL_DIR}/${BUILDPACK_BUILD}

install_all: install_sqlbundle install_buildpack

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

all: clean build_all compress-all
