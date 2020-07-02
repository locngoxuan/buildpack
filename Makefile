.PHONY: buildpack

BUILD=buildpack
INSTALL_DIR=/usr/local/bin

dev:
	env CGO_ENABLED=0 go build -ldflags="-s -w" -o ./bin/${BUILD} -a ./cmd

install:
	chmod 755 ./bin/${BUILD}
	cp -r ./bin/${BUILD} ${INSTALL_DIR}/${BUILD}

clean:
	rm -rf ./bin