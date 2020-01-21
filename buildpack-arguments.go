package main

import (
	"flag"
	"os"
	"strings"
)

type ActionArguments struct {
	Flag   *flag.FlagSet
	Values map[string]interface{}
}

func newActionArguments(f *flag.FlagSet) *ActionArguments {
	return &ActionArguments{
		Flag:   f,
		Values: make(map[string]interface{}),
	}
}

func (a *ActionArguments) readVersion() *ActionArguments {
	s := a.Flag.String("v", "", "version number")
	a.Values["v"] = s
	return a
}

func (a *ActionArguments) readModules() *ActionArguments {
	s := a.Flag.String("m", "", "modules")
	a.Values["m"] = s
	return a
}

func (a *ActionArguments) parse() error {
	return a.Flag.Parse(os.Args[2:])
}

func (a *ActionArguments) version() string {
	s, ok := a.Values["v"]
	if !ok {
		return ""
	}
	return strings.TrimSpace(*(s.(*string)))
}

func (a *ActionArguments) modules() []string {
	s, ok := a.Values["m"]
	if !ok {
		return []string{}
	}
	v := strings.TrimSpace(*(s.(*string)))
	if len(v) == 0 {
		return []string{}
	}
	return strings.Split(v, ",")
}

func (a *ActionArguments) readContainer() *ActionArguments {
	s := a.Flag.Bool("container", false, "using docker environment rather than host environment")
	a.Values["container"] = s
	return a
}

func (a *ActionArguments) container() bool {
	s, ok := a.Values["container"]
	if !ok {
		return false
	}
	return *(s.(*bool))
}
