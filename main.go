package main

import (
	"fmt"
	"github.com/johnnylei/my_docker/subsystem"
	"github.com/johnnylei/my_docker/util"
	"github.com/urfave/cli"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

//func main() {
//	var wg sync.WaitGroup
//	wg.Add(2)
//	_, write, _ := util.NewPipe()
//	go func() {
//		write.WriteString("hello fucker")
//		wg.Done()
//	}()
//
//	go func() {
//		buffer := make([]byte, 1024)
//		reader := os.NewFile(uintptr(3), "pipe")
//		reader.Read(buffer)
//		fmt.Println(string(buffer))
//		wg.Done()
//	}()
//	wg.Wait()
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

				read, write, err := util.NewPipe()
				if err != nil {
					_, file, line, _ := runtime.Caller(1)
					fmt.Printf("file%s, line:%d, err:%s\n", file, line, err.Error())
					return err
				}

				cmd := exec.Command("/proc/self/exe", "init")
				cmd.SysProcAttr = &syscall.SysProcAttr{
					Cloneflags:syscall.CLONE_NEWIPC | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
				}

				if tty := c.Bool("ti"); tty {
					cmd.Stdin = os.Stdin
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
				}

				cmd.ExtraFiles = append(cmd.ExtraFiles, read)
				if _, err := write.WriteString(strings.Join(c.Args(), " ")); err != nil {
					_, file, line, _ := runtime.Caller(1)
					fmt.Printf("file%s, line:%d, err:%s\n", file, line, err.Error())
					return err
				}

				if err := cmd.Start(); err != nil {
					_, file, line, _ := runtime.Caller(1)
					fmt.Printf("file%s, line:%d, err:%s\n", file, line, err.Error())
					return err
				}

				resourceConfig := &subsystem.ResourceConfig {
					MemoryLimit: c.Int("m"),
					CpuSet: c.String("cpuset"),
					CpuShare: c.String("cpushare"),
				}

				manager, err := subsystem.InitCgroupsManager("mydocker-cgroup", resourceConfig)
				if err != nil {
					_, file, line, _ := runtime.Caller(1)
					fmt.Printf("file%s, line:%d, err:%s\n", file, line, err.Error())
					return err
				}

				if err := manager.Run(cmd.Process.Pid); err != nil {
					_, file, line, _ := runtime.Caller(1)
					fmt.Printf("file%s, line:%d, err:%s\n", file, line, err.Error())
					return err
				}

				defer manager.Destroy()

				if err := cmd.Wait(); err != nil {
					_, file, line, _ := runtime.Caller(1)
					fmt.Printf("file%s, line:%d, err:%s\n", file, line, err.Error())
					return err
				}

				return nil
			},
		},
		{
			Name: "init",
			Usage: "init for container",
			Action: func(c *cli.Context) error {
				if err := syscall.Mount("proc", "/proc", "proc", uintptr(syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NOSUID), ""); err != nil {
					_, file, line, _ := runtime.Caller(1)
					fmt.Printf("file%s, line:%d, err:%s\n", file, line, err.Error())
					return err
				}

				reader := os.NewFile(uintptr(3), "pipe")
				_, message, err := util.ReadPipe(reader)
				if err != nil {
					_, file, line, _ := runtime.Caller(1)
					fmt.Printf("file%s, line:%d, err:%s\n", file, line, err.Error())
					return err
				}

				command := strings.Split(message, " ")
				path, err := exec.LookPath(command[0])
				if err != nil {
					_, file, line, _ := runtime.Caller(1)
					fmt.Printf("file%s, line:%d, err:%s\n", file, line, err.Error())
					return err
				}

				if err:= syscall.Exec(path, command[0:], os.Environ()); err !=nil {
					_, file, line, _ := runtime.Caller(1)
					fmt.Printf("file%s, line:%d, err:%s\n", file, line, err.Error())
					return err
				}
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}