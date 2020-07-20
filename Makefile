.PHONY: buildpack

BUILD=buildpack
INSTALL_DIR=/usr/local/bin
VERSION=2.0.0
BUILD_ID=1

dev:
	@export GOPROXY=direct
	@export GOSUMDB=off
	go get -v .
	env CGO_ENABLED=1 go build -ldflags="-s -w -X main.version=${VERSION}" -o ./bin/${BUILD} -a ./cmd
	@sleep 1

install:
	chmod 755 ./bin/${BUILD}
	cp -r ./bin/${BUILD} ${INSTALL_DIR}/${BUILD}

### RPM BUILD
rpm: rpm_prepare_workspace rpm_prepare_source rpm_build

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

rpm_build:
	@echo "Building rmp..."
	rpmbuild --define "_topdir `pwd`/rpmbuild"  --define "BUILD_ID $(BUILD_ID)"  --define "BUILD_VERSION $(VERSION)" -ba rpmbuild/SPECS/buildpack.spec
	@rm -rf rpmbuild/SOURCES/buildpack*.tar.gz

clean:
	rm -rf bin
	rm -rf rpmbuild
	@sleep 1