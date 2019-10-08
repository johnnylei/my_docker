package main

import (
	"fmt"
	"github.com/johnnylei/my_docker/subsystem"
	"github.com/urfave/cli"
	"log"
	"os"
	"os/exec"
	"syscall"
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
				cli.IntFlag{
					Name: "m",
				},
				cli.StringFlag{
					Name: "cpuset",
				},
				cli.StringFlag{
					Name: "cpushare",
				},
			},
			Action: func(c *cli.Context) error {
				if len(c.Args()) < 1 {
					return fmt.Errorf("missing container command")
				}

				cmd := exec.Command("/proc/self/exe", "init", c.Args().Get(0))
				cmd.SysProcAttr = &syscall.SysProcAttr{
					Cloneflags:syscall.CLONE_NEWIPC | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
				}

				if tty := c.Bool("ti"); tty {
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
				}

				if err := cmd.Start(); err != nil {
					log.Fatal(err)
				}

				resourceConfig := &subsystem.ResourceConfig{
					MemoryLimit: c.Int("m"),
					CpuSet: c.String("cpuset"),
					CpuShare: c.String("cpushare"),
				}

				manager, err := subsystem.InitCgroupsManager("mydocker-cgroup", resourceConfig)
				if err != nil {
					log.Fatal(err)
				}

				if err := manager.Run(); err != nil {
					log.Fatal(err)
				}

				defer manager.Destroy()

				if err := cmd.Wait(); err != nil {
					log.Fatal(err)
				}

				return nil
			},
		},
		{
			Name: "init",
			Usage: "init for container",
			Action: func(c *cli.Context) error {
				command := c.Args().Get(0)
				if err := syscall.Mount("proc", "/proc", "proc", uintptr(syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NOSUID), ""); err != nil {
					log.Fatal(err)
				}

				argv := []string{command}
				if err:= syscall.Exec(command, argv, os.Environ()); err !=nil {
					log.Fatal(err)
				}
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}