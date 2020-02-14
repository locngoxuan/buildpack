package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"scm.wcs.fortna.com/lngo/buildpack/sqlbundle"
	"strings"
)

var (
	flags *flag.FlagSet
)

func init() {
	flags = flag.NewFlagSet("sqlbundle", flag.ContinueOnError)
}

func main() {
	configFile := flags.String("config", "", "path to specific configuration file")
	clean := flags.Bool("clean", false, "clean after build")

	err := flags.Parse(os.Args[1:])
	if err != nil {
		flags.Usage()
		os.Exit(2)
	}

	var root string
	var bundleFile string
	if len(strings.TrimSpace(*configFile)) > 0 {
		bundleFile = strings.TrimSpace(*configFile)
		root, _ = filepath.Split(bundleFile)
	} else {
		root, err = filepath.Abs(".")
		if err != nil {
			panic(err)
		}
		bundleFile = filepath.Join(root, sqlbundle.FileConfig)
	}

	bundle := sqlbundle.SQLBundle{
		WorkingDir: root,
		BundleFile: bundleFile,
		Clean:      *clean,
	}
	err = bundle.Run(os.Stdout)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
