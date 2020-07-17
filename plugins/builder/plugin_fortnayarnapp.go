package main

import "scm.wcs.fortna.com/lngo/buildpack/builder"

type YarnApp struct {
	builder.Yarn
}

func YarnAppBuilder() builder.Interface {
	return &YarnApp{}
}