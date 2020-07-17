package main

import (
	"scm.wcs.fortna.com/lngo/buildpack/builder"
)

func YarnLibBuilder() builder.Interface {
	return &YarnLib{}
}

type YarnLib struct {
	builder.Yarn
}
