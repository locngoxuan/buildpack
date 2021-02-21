package main

const (
	DockerfileOfBuilder = `#generated dockerfile
FROM alpine:3.13.2 as builder
MAINTAINER Buildpack <xuanloc0511@gmail.com>
RUN mkdir -p /working
ADD . /working

FROM {{.Image}}
MAINTAINER Buildpack <xuanloc0511@gmail.com>
RUN mkdir -p /working
ADD . /working
WORKDIR /working
`
)

type BuilderTemplate struct {
	Image   string
	Command string
}
