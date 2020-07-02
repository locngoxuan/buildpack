package buildpack

import "flag"

var f *flag.FlagSet

type Arguments struct {
	Version string
}