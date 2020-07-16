package main

import "scm.wcs.fortna.com/lngo/buildpack/builder"

type YarnApp struct {
	builder.Yarn
}

func init() {
	registries["yarn_app"] = &YarnApp{}
}
