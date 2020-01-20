package main

import (
	"flag"
	"strings"
)

func readVersion(f *flag.FlagSet) string {
	s := f.String("v", "0.1.0", "version number")
	return strings.TrimSpace(*s)
}

func readModules(f *flag.FlagSet) []string {
	s := f.String("m", "", "modules")
	if len(strings.TrimSpace(*s)) == 0 {
		return []string{}
	}
	return strings.Split(*s, ",")
}

func readContainerOpt(f *flag.FlagSet) bool {
	s := f.Bool("container", false, "using docker environment rather than host environment")
	return *s
}

