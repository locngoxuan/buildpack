package main

import (
	"flag"
	"strings"
)

func readVersion(f *flag.FlagSet) string {
	s := f.String("v", "0.1.0", "version number")
	return strings.TrimSpace(*s)
}
