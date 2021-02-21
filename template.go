package main

const (
	DockerfileOfBuilder = `#generated dockerfile
FROM {{.Image}}
MAINTAINER Buildpack <xuanloc0511@gmail.com>
RUN mkdir -p /working
ADD . /working
WORKDIR /working
`
)

type BuilderTemplate struct {
	Image   string
}
