package main

import (
	"fmt"
	"github.com/johnnylei/my_docker/network"
	"net"
)

func main() {
	ipam := network.IPAM{
		SubnetAllocatedPath: "/Users/leiyuqing02/go/src/github.com/johnnylei/my_docker/ipam.json",
		Subnets:&map[string]string{},
	}




}

//func main() {
//	app := cli.NewApp()
//	app.Commands = []cli.Command{
//		{
//			Name: "run",
//			Usage: `create a container with namespace and cgroups limit mydocker run -ti [command]`,
//			Flags: []cli.Flag{
//				cli.BoolFlag{
//					Name: "ti",
//				},
//				cli.BoolFlag{
//					Name: "d",
//				},
//				cli.IntFlag{
//					Name: "m",
//				},
//				cli.StringFlag{
//					Name: "cpuset",
//				},
//				cli.StringFlag{
//					Name: "cpushare",
//				},
//				cli.StringFlag{
//					Name: "name",
//				},
//				cli.StringSliceFlag{
//					Name: "env",
//					Usage: "set environment",
//				},
//				cli.StringFlag{
//					Name: "image",
//					Value: "busybox",
//				},
//				cli.StringFlag{
//					Name: "v",
//					Usage: "volume mount",
//				},
//			},
//			Action: func(c *cli.Context) error {
//				return container.Run(c)
//			},
//		},
//		{
//			Name: "init",
//			Usage: "init for container",
//			Flags: []cli.Flag{
//				cli.StringFlag{
//					Name: "name",
//				},
//				cli.StringFlag{
//					Name: "v",
//					Usage: "volume",
//				},
//				cli.StringFlag{
//					Name: "root",
//					Usage: "root path",
//				},
//			},
//			Action: func(c *cli.Context) error {
//				return container.Init(c)
//			},
//		},
//		{
//			Name: "delete",
//			Flags:[]cli.Flag{
//				cli.StringFlag{
//					Name: "name",
//				},
//			},
//			Action: func(c *cli.Context) error {
//				return container.Delete(c)
//			},
//		},
//		{
//			Name: "commit",
//			Flags:[]cli.Flag{
//				cli.StringFlag{
//					Name: "name",
//				},
//			},
//			Action: func(context *cli.Context) error {
//				return container.Commit(context)
//			},
//		},
//		{
//			Name: "ps",
//			Action: func(context *cli.Context) error {
//				return container.Ps(context)
//			},
//		},
//		{
//			Name: "logs",
//			Flags:[]cli.Flag{
//				cli.StringFlag{
//					Name: "name",
//				},
//			},
//			Action: func(context *cli.Context) error {
//				return container.Logs(context)
//			},
//		},
//		{
//			Name: "stop",
//			Flags:[]cli.Flag{
//				cli.StringFlag{
//					Name: "name",
//				},
//			},
//			Action: func(context *cli.Context) error {
//				return container.Stop(context)
//			},
//		},
//		{
//			Name: "exec",
//			Flags:[]cli.Flag{
//				cli.StringFlag{
//					Name: "name",
//				},
//				cli.BoolFlag{
//					Name: "ti",
//				},
//				cli.BoolFlag{
//					Name: "d",
//				},
//				cli.BoolFlag{
//					Name: "child",
//				},
//			},
//			Action: func(context *cli.Context) error {
//				return container.Exec(context)
//			},
//		},
//		{
//			Name:"image",
//			Subcommands:[]cli.Command{
//				{
//					Name:"ps",
//					Action: func(context *cli.Context)error {
//						return image.Ps(context)
//					},
//				},
//			},
//		},
//	}
//
//	if err := app.Run(os.Args); err != nil {
//		log.Fatal(err)
//	}
//}