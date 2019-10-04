package main

import (
	"github.com/johnnylei/my_docker/cgroups"
	"github.com/johnnylei/my_docker/namespace"
)

func main() {
	namespace.Segregate()
	cgroups.Run()
}