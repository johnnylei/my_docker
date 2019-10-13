package main

import (
	"github.com/johnnylei/my_docker/container"
	"github.com/urfave/cli"
	"log"
	"os"
)

//
//func main() {
//	containerInformation := &container.ContainerInformation{
//		Pid: 1,
//		Id: "xxx",
//		Name: "namexxx",
//		InitCommand: "init command",
//		Status: container.STATUS_RUNING,
//		CreatedTime: time.Now().Format("2006-01-02 15:04:05"),
//	}
//	InformationBytes, _ := json.Marshal(containerInformation)
//	log.Println(string(InformationBytes))
//}

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
				},
				cli.StringFlag{
					Name: "v",
					Usage: "volume mount",
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
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}