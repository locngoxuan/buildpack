.PHONY: clean download_maven build

VERSION=3.6.3

clean:
	rm -rf tmp

download_maven: clean
	mkdir -p tmp
	curl https://archive.apache.org/dist/maven/maven-3/$(VERSION)/binaries/apache-maven-$(VERSION)-bin.tar.gz -o tmp/apache-maven-$(VERSION)-bin.tar.gz
	tar zxvf tmp/apache-maven-$(VERSION)-bin.tar.gz -C tmp/
	mv tmp/apache-maven-$(VERSION) tmp/maven

build: download_maven
	docker build . --no-cache --tag $(DOCKER_TAG)