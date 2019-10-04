package main

import (
	"github.com/johnnylei/my_docker/cgroups"
)

func main() {
	//namespace.Segregate()
	cgroups.Run()
}