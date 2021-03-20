package main

import "github.com/locngoxuan/buildpack"

var version = "2.1.2"

func main()  {
	buildpack.SetVersion(version)
	buildpack.Run()
}
