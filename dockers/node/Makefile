.PHONY: docker

DOCKER_TAG?=lts-alpine3.13
NODE?=alphine313

docker:
	docker build . --no-cache --tag xuanloc0511/node:lts-alpine3.13 -f Dockerfile.$(NODE)

