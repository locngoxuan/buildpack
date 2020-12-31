.PHONY: buildpack

BUILD=buildpack
INSTALL_DIR=/usr/local/bin
VERSION=1.1.2
BUILD_ID=1
BUILD_OS=linux
BUILD_ARCH=amd64

dev:
	@export GOPROXY=direct
	@export GOSUMDB=off
	go get -v .
	env CGO_ENABLED=1 go build -ldflags="-s -w -X main.version=${VERSION}" -o ./bin/${BUILD} -a ./cmd

build:
	@export GOPROXY=direct
	@export GOSUMDB=off
	go get -v .
	env GOOS=${BUILD_OS} GOARCH=${BUILD_ARCH} CGO_ENABLED=1 go build -ldflags="-s -w -X main.version=${VERSION}" -o ./bin/${BUILD} -a ./cmd
	@sleep 1

install:
	chmod 755 ./bin/${BUILD}
	cp -r ./bin/${BUILD} ${INSTALL_DIR}/${BUILD}

### RPM BUILD
rpm: build rpm_prepare_workspace rpm_prepare_source rpm_build

rpm_mac: build rpm_prepare_workspace rpm_prepare_source_mac rpm_build

rpm_prepare_workspace:
	@echo "Prepare directories for RPM building"
	mkdir -p rpmbuild && mkdir -p rpmbuild/BUILD \
		&& mkdir -p rpmbuild/RPMS \
		&& mkdir -p rpmbuild/SOURCES \
		&& mkdir -p rpmbuild/SPECS \
		&& mkdir -p rpmbuild/SRPMS
	@sleep 1

rpm_prepare_source:
	@echo "Prepare sources are needed for RPM building"
	cp -rf buildpack.spec rpmbuild/SPECS/buildpack.spec
	@rm -rf rpmbuild/SOURCES/buildpack*.tar.gz
	tar -czvf buildpack-$(VERSION).tar.gz --transform s/^bin/buildpack-$(VERSION)/ bin
	mv buildpack-$(VERSION).tar.gz rpmbuild/SOURCES/
	@sleep 1

rpm_prepare_source_mac:
	@echo "Prepare sources are needed for RPM building"
	cp -rf buildpack.spec rpmbuild/SPECS/buildpack.spec
	@rm -rf rpmbuild/SOURCES/buildpack*.tar.gz
	tar -czvf buildpack-$(VERSION).tar.gz -s /^bin/buildpack-$(VERSION)/ bin
	mv buildpack-$(VERSION).tar.gz rpmbuild/SOURCES/
	@sleep 1

rpm_build:
	@echo "Building rpm..."
	rpmbuild --define "_topdir `pwd`/rpmbuild"  \
		--define "BUILD_ID $(BUILD_ID)"  \
		--define "BUILD_VERSION $(VERSION)" \
		--define "BUILD_OS $(BUILD_OS)" \
		-ba rpmbuild/SPECS/buildpack.spec
	@rm -rf rpmbuild/SOURCES/buildpack*.tar.gz

clean:
	rm -rf bin
	rm -rf rpmbuild
	@sleep 1
