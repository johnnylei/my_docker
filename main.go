package main

import (
	"github.com/johnnylei/my_docker/container"
	"github.com/johnnylei/my_docker/image"
	"github.com/johnnylei/my_docker/network"
	"github.com/urfave/cli"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name: "run",
			Usage: `create a container with namespace and cgroups limit mydocker run -ti [command]`,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "ti",
				},
				cli.BoolFlag{
					Name: "d",
				},
				cli.IntFlag{
					Name: "m",
				},
				cli.StringFlag{
					Name: "cpuset",
				},
				cli.StringFlag{
					Name: "cpushare",
				},
				cli.StringFlag{
					Name: "name",
					Required:true,
				},
				cli.StringSliceFlag{
					Name: "env",
					Usage: "set environment",
				},
				cli.StringFlag{
					Name: "image",
					Value: "busybox",
				},
				cli.StringFlag{
					Name: "v",
					Usage: "volume mount",
				},
				cli.StringFlag{
					Name: "net",
					Value: "mydocker0",
				},
				cli.StringSliceFlag{
					Name: "p",
					Usage: "port map",
				},
			},
			Action: func(c *cli.Context) error {
				return container.Run(c)
			},
		},
		{
			Name: "init",
			Usage: "init for container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "name",
				},
				cli.StringFlag{
					Name: "v",
					Usage: "volume",
				},
				cli.StringFlag{
					Name: "root",
					Usage: "root path",
				},
			},
			Action: func(c *cli.Context) error {
				return container.Init(c)
			},
		},
		{
			Name: "delete",
			Flags:[]cli.Flag{
				cli.StringFlag{
					Name: "name",
				},
			},
			Action: func(c *cli.Context) error {
				return container.Delete(c)
			},
		},
		{
			Name: "commit",
			Flags:[]cli.Flag{
				cli.StringFlag{
					Name: "name",
				},
			},
			Action: func(context *cli.Context) error {
				return container.Commit(context)
			},
		},
		{
			Name: "ps",
			Action: func(context *cli.Context) error {
				return container.Ps(context)
			},
		},
		{
			Name: "logs",
			Flags:[]cli.Flag{
				cli.StringFlag{
					Name: "name",
				},
			},
			Action: func(context *cli.Context) error {
				return container.Logs(context)
			},
		},
		{
			Name: "stop",
			Flags:[]cli.Flag{
				cli.StringFlag{
					Name: "name",
				},
			},
			Action: func(context *cli.Context) error {
				return container.Stop(context)
			},
		},
		{
			Name: "exec",
			Flags:[]cli.Flag{
				cli.StringFlag{
					Name: "name",
				},
				cli.BoolFlag{
					Name: "ti",
				},
				cli.BoolFlag{
					Name: "d",
				},
				cli.BoolFlag{
					Name: "child",
				},
			},
			Action: func(context *cli.Context) error {
				return container.Exec(context)
			},
		},
		{
			Name:"image",
			Subcommands:[]cli.Command{
				{
					Name:"ps",
					Action: func(context *cli.Context)error {
						return image.Ps(context)
					},
				},
			},
		},
		{
			Name:"network",
			Usage:"network manager",
			Subcommands:[]cli.Command{
				{
					Name: "driver",
					Usage: "network driver manger",
					Subcommands: []cli.Command{
						{
							Name:"create",
							Usage:"mydocker network driver create -name xxx -subnet xxxx",
							Flags:[]cli.Flag{
								cli.StringFlag{
									Name:"name",
									Required:true,
								},
								cli.StringFlag{
									Name:"subnet",
									Required:true,
								},
							},
							Action: func(context *cli.Context) error {
								return network.CreateBridgeInterface(context)
							},
						},
						{
							Name: "delete",
							Usage: "mydocker network driver delete --name xxx",
							Flags:[]cli.Flag{
								cli.StringFlag{
									Name:"name",
									Required:true,
								},
							},
							Action: func(context *cli.Context) error {
								return network.DeleteBridgeInterface(context)
							},
						},
					},
				},
				{
					Name:"create",
					Usage:"my_docker create --subnet 172.17.0.0/16 --driver xxx --name xxx",
					Flags:[]cli.Flag{
						cli.StringFlag{
							Name:"name",
							Required:true,
						},
						cli.StringFlag{
							Name:"driver-type",
							Required:true,
						},
						cli.StringFlag{
							Name:"subnet",
							Value:"172.17.0.0/16",
						},
					},
					Action: func(context *cli.Context) error {
						return network.CreateNetwork(context)
					},
				},
				{
					Name:"delete",
					Usage:"my_docker delete --name xxx",
					Flags:[]cli.Flag{
						cli.StringFlag{
							Name:"name",
							Required:true,
						},
					},
					Action: func(context *cli.Context) error {
						return network.DeleteNetwork(context)
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}