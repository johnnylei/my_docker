package main

import (
	"github.com/johnnylei/my_docker/base/cgroups"
)

func main() {
	//namespace.Segregate()
	cgroups.Run()
}