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
	dockerize := flags.Bool("dockerize", false, "build docker image")
	configFile := flags.String("config", "", "path to specific configuration file")
	clean := flags.Bool("clean", false, "clean after build")

	err := flags.Parse(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		flags.Usage()
		os.Exit(2)
	}

	var root string
	var bundleFile string
	if len(strings.TrimSpace(*configFile)) > 0 {
		bundleFile = strings.TrimSpace(*configFile)
		root, err = filepath.Abs(strings.TrimSpace(*configFile))
		if err != nil {
			panic(err)
		}
	} else {
		root, err = filepath.Abs(".")
		if err != nil {
			panic(err)
		}
		bundleFile = filepath.Join(root, "sqlbundle.yml")
	}

	bundle := sqlbundle.SQLBundle{
		WorkingDir:  root,
		BundleFile:  bundleFile,
		Clean:       *clean,
		Dockerize:   *dockerize,
		DockerHosts: []string{"unix:///var/run/docker.sock", "tcp://127.0.0.1:2375"},
	}
	err = bundle.Run(os.Stdout)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
