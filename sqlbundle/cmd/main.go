package main

import (
	"fmt"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/sqlbundle"
)

func main() {
	root, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	bundle := sqlbundle.SQLBundle{
		WorkingDir: root,
	}
	err = bundle.Run()
	fmt.Println(err)
}
