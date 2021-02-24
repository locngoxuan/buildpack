package main

import (
	"bytes"
	"fmt"
	"text/template"
)

const (
	DockerfileOfBuilder = `#generated dockerfile
FROM {{.Image}}
MAINTAINER Buildpack <xuanloc0511@gmail.com>
RUN mkdir -p /working
ADD . /working
WORKDIR /working
`

	ErrorDetail = `{{.Error}}
{{.Detail}}`
)

type BuilderTemplate struct {
	Image string
}

func fmtError(err error, msg string) error {
	type ErrTemp struct {
		Error  string
		Detail string
	}
	t := template.Must(template.New("error").Parse(ErrorDetail))
	var buf bytes.Buffer
	defer buf.Reset()
	e := t.Execute(&buf, ErrTemp{
		Error:  err.Error(),
		Detail: msg,
	})
	if e != nil {
		return err
	}
	return fmt.Errorf(buf.String())
}
